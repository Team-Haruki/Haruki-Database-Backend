from typing import Tuple
from datetime import datetime, UTC
from pydantic import ValidationError
from sqlalchemy import select, delete, and_
from quart import Blueprint, request, Response
import orjson

from api.utils import success, error, redis_client
from modules.sql.tables.pjsk import (
    Alias,
    PendingAlias,
    RejectedAlias,
    GroupAlias,
    AliasAdmin,
)
from ..db_engine import engine
from ..schema import AliasBodySchema, AliasApprovalSchema, AliasRejectionSchema, PendingAliasSchema

alias_api = Blueprint("pjsk_alias", __name__, url_prefix="/alias")


async def is_alias_admin(im_id: str) -> bool:
    async with engine.session() as session:
        result = await session.execute(select(AliasAdmin).where(AliasAdmin.im_id == im_id))
        return result.scalar_one_or_none() is not None


@alias_api.get("/<alias_type>-id")
async def get_alias_type_id(alias_type: str) -> Tuple[Response, int]:
    alias = request.args.get("alias")
    group_id = request.args.get("group_id", None)
    if not alias:
        return error("Missing alias parameter", code=400)
    if alias_type not in {"music", "character"}:
        return error("Invalid alias type", code=400)
    try:
        async with engine.session() as session:
            if group_id:
                stmt = select(GroupAlias.alias_type_id).where(
                    and_(
                        GroupAlias.alias_type == alias_type, GroupAlias.alias == alias, GroupAlias.group_id == group_id
                    )
                )
            else:
                stmt = select(Alias.alias_type_id).where(and_(Alias.alias_type == alias_type, Alias.alias == alias))
            result = await session.execute(stmt)
            rows = result.scalars().all()
            if not rows:
                return error("Alias not found", code=404)
            return success({"target_ids": rows})
    except Exception as e:
        return error(f"Internal server error: {str(e)}", code=500)


@alias_api.get("/<alias_type>/<alias_type_id>")
async def get_alias(alias_type: str, alias_type_id: str) -> Tuple[Response, int]:
    if alias_type not in {"music", "character"}:
        return error("Invalid alias type", code=400)
    try:
        cache_key = f"aliases:{alias_type}:{alias_type_id}"
        cached = await redis_client.get(cache_key)
        if cached:
            return success(orjson.loads(cached))

        async with engine.session() as session:
            stmt = select(Alias.alias).where(and_(Alias.alias_type == alias_type, Alias.alias_type_id == alias_type_id))
            result = await session.execute(stmt)
            aliases = result.scalars().all()
            await redis_client.set(cache_key, orjson.dumps(aliases))
            return success(aliases)
    except Exception as e:
        return error(f"Internal server error: {str(e)}", code=500)


@alias_api.post("/<alias_type>/<alias_type_id>/add")
async def add_alias(alias_type: str, alias_type_id: str) -> Tuple[Response, int]:
    try:
        data = AliasBodySchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())

    if alias_type not in {"music", "character"}:
        return error("Invalid alias type", code=400)

    if not await is_alias_admin(data.im_id):
        async with engine.session() as session:
            session.add(
                PendingAlias(
                    alias_type=alias_type,
                    alias_type_id=int(alias_type_id),
                    alias=data.alias,
                    submitted_by=data.im_id,
                    submitted_at=datetime.now(UTC),
                )
            )
            await session.commit()
        return success(message="Alias submitted for review.", code=202)

    async with engine.session() as session:
        session.add(Alias(alias_type=alias_type, alias_type_id=int(alias_type_id), alias=data.alias))
        await session.commit()
    await redis_client.delete(f"aliases:{alias_type}:{alias_type_id}")

    return success(message="Alias added.", code=201)


@alias_api.delete("/<alias_type>/<alias_type_id>/<internal_id>")
async def remove_alias(alias_type: str, alias_type_id: str, internal_id: str) -> Tuple[Response, int]:
    try:
        data = AliasBodySchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())

    if alias_type not in {"music", "character"}:
        return error("Invalid alias type", code=400)

    if not await is_alias_admin(data.im_id):
        return error("Permission denied.", code=403)

    async with engine.session() as session:
        await session.execute(
            delete(Alias).where(
                and_(
                    Alias.alias_type == alias_type,
                    Alias.alias_type_id == int(alias_type_id),
                    Alias.alias == data.alias,
                    Alias.id == int(internal_id),
                )
            )
        )
        await session.commit()
    await redis_client.delete(f"aliases:{alias_type}:{alias_type_id}")
    return success(message="Alias deleted.")


@alias_api.get("/pending")
async def get_pending_alias() -> Tuple[Response, int]:
    im_id = request.args.get("im_id")
    if not im_id:
        return error("Missing im_id", code=400)
    if not await is_alias_admin(im_id):
        return error("Permission denied", code=403)
    async with engine.session() as session:
        stmt = select(PendingAlias)
        result = await session.execute(stmt)
        rows = result.scalars().all()
        if not rows:
            return error("No pending review alias.", code=404)
        return success([PendingAliasSchema.model_validate(row).model_dump() for row in rows])


@alias_api.post("/pending/<pending_id>/approve")
async def approve_alias(pending_id: str) -> Tuple[Response, int]:
    try:
        data = AliasApprovalSchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())

    if not await is_alias_admin(data.im_id):
        return error("Permission denied.", code=403)

    async with engine.session() as session:
        result = await session.execute(select(PendingAlias).where(PendingAlias.id == int(pending_id)))
        row = result.scalar_one_or_none()
        if not row:
            return error("Pending alias not found", code=404)

        session.add(Alias(alias_type=row.alias_type, alias_type_id=row.alias_type_id, alias=row.alias))
        await session.delete(row)
        await session.commit()
    return success("Alias approved and added.", code=201)


@alias_api.post("/pending/<pending_id>/reject")
async def reject_alias(pending_id: str) -> Tuple[Response, int]:
    try:
        data = AliasRejectionSchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())

    if not await is_alias_admin(data.im_id):
        return error("Permission denied.", code=403)

    async with engine.session() as session:
        result = await session.execute(select(PendingAlias).where(PendingAlias.id == int(pending_id)))
        row = result.scalar_one_or_none()
        if not row:
            return error("Pending alias not found", code=404)

        session.add(
            RejectedAlias(
                id=row.id,
                alias_type=row.alias_type,
                alias_type_id=row.alias_type_id,
                alias=row.alias,
                reviewed_by=data.im_id,
                reason=data.reason,
                reviewed_at=datetime.now(UTC),
            )
        )
        await session.delete(row)
        await session.commit()
    return success("Alias rejected and logged.", code=201)


@alias_api.get("/status/<pending_id>")
async def get_alias_review_status(pending_id: str) -> Tuple[Response, int]:
    async with engine.session() as session:
        result = await session.execute(select(PendingAlias).where(PendingAlias.id == int(pending_id)))
        pending = result.scalar_one_or_none()
        if pending:
            return success(message="This alias is pending review.", code=200)

        result = await session.execute(select(RejectedAlias.reason).where(RejectedAlias.id == int(pending_id)))
        reason = result.scalar_one_or_none()
        if reason:
            return error(message=reason, code=400)

    return error("Alias review record not found.", code=404)


@alias_api.get("/group/<group_id>/<alias_type>/<alias_type_id>")
async def get_group_alias(group_id: str, alias_type: str, alias_type_id: str) -> Tuple[Response, int]:
    if alias_type not in {"music", "character"}:
        return error("Invalid alias type", code=400)
    try:
        cache_key = f"group_aliases:{group_id}:{alias_type}:{alias_type_id}"
        cached = await redis_client.get(cache_key)
        if cached:
            return success(orjson.loads(cached))
        async with engine.session() as session:
            stmt = select(GroupAlias.alias).where(
                and_(
                    GroupAlias.group_id == group_id,
                    GroupAlias.alias_type == alias_type,
                    GroupAlias.alias_type_id == int(alias_type_id),
                )
            )
            result = await session.execute(stmt)
            aliases = result.scalars().all()
            if not aliases:
                return error("No aliases found for this group", code=404)
            await redis_client.set(cache_key, orjson.dumps(aliases))
            return success(aliases)
    except Exception as e:
        return error(f"Internal server error: {str(e)}", code=500)


@alias_api.post("/group/<group_id>/<alias_type>/<alias_type_id>")
async def add_group_alias(group_id: str, alias_type: str, alias_type_id: str) -> Tuple[Response, int]:
    try:
        data = AliasBodySchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())

    if alias_type not in {"music", "character"}:
        return error("Invalid alias type", code=400)

    async with engine.session() as session:
        session.add(
            GroupAlias(group_id=group_id, alias_type=alias_type, alias_type_id=int(alias_type_id), alias=data.alias)
        )
        await session.commit()
    await redis_client.delete(f"group_aliases:{group_id}:{alias_type}:{alias_type_id}")

    return success(message="Group alias added.", code=201)


@alias_api.delete("/group/<group_id>/<alias_type>/<alias_type_id>")
async def remove_group_alias(group_id: str, alias_type: str, alias_type_id: str) -> Tuple[Response, int]:
    try:
        data = AliasBodySchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())

    if alias_type not in {"music", "character"}:
        return error("Invalid alias type", code=400)

    async with engine.session() as session:
        await session.execute(
            delete(GroupAlias).where(
                and_(
                    GroupAlias.group_id == group_id,
                    GroupAlias.alias_type == alias_type,
                    GroupAlias.alias_type_id == int(alias_type_id),
                    GroupAlias.alias == data.alias,
                )
            )
        )
        await session.commit()
    await redis_client.delete(f"group_aliases:{group_id}:{alias_type}:{alias_type_id}")

    return success(message="Group alias deleted.")
