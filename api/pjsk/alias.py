from sqlalchemy import and_
from typing import Optional, Union
from datetime import datetime, UTC
from fastapi import APIRouter, Query, Depends

from utils import pjsk_engine as engine
from utils import redis_client, is_alias_admin, require_alias_admin, parse_json_body, get_cached_data, verify_api_auth
from modules.enums import AliasType
from modules.schemas.response import APIResponse
from modules.schemas.pjsk import (
    AliasSchema,
    AllAliasesSchema,
    PendingAliasList,
    PendingAliasEntry,
    AliasApprovalSchema,
    AliasRejectionSchema,
    AliasToObjectIdSchema,
)
from modules.sql.tables.pjsk import (
    Alias,
    GroupAlias,
    PendingAlias,
    RejectedAlias,
)

alias_api = APIRouter(prefix="/alias", tags=["Alias"])


@alias_api.get(
    "/{alias_type}-id",
    response_model=Union[AliasToObjectIdSchema, APIResponse],
    summary="根据别名获取目标类型ID",
    description="根据歌曲/角色别名返回所有对应类型ID",
)
async def get_alias_type_id(
    alias_type: AliasType,
    alias: str = Query(..., description="Alias to lookup"),
    group_id: Optional[str] = Query(None, description="Optional group ID"),
) -> Union[AliasToObjectIdSchema, APIResponse]:
    try:
        if group_id:
            select_object, select_clause = (
                GroupAlias.alias_type_id,
                and_(GroupAlias.alias_type == alias_type, GroupAlias.alias == alias, GroupAlias.group_id == group_id),
            )
        else:
            select_object, select_clause = (
                Alias.alias_type_id,
                and_(Alias.alias_type == alias_type, Alias.alias == alias),
            )
        rows = await engine.select_data(select_object, select_clause)
        if not rows:
            return APIResponse(status=404, message="Alias not found")
        return AliasToObjectIdSchema(match_ids=rows)
    except Exception as e:
        return APIResponse(status=500, message=f"Internal server error: {str(e)}")


@alias_api.get(
    "/{alias_type}/{alias_type_id}",
    response_model=Union[AllAliasesSchema, APIResponse],
    summary="获取指定别名与指定别名类型ID的全部别名",
    description="根据歌曲/角色ID返回所有对应别名",
)
async def get_alias(alias_type: AliasType, alias_type_id: int) -> Union[AllAliasesSchema, APIResponse]:
    try:
        cache_key = f"pjsk-aliases:{alias_type}:{alias_type_id}"
        aliases = await get_cached_data(
            engine, cache_key, Alias.alias, and_(Alias.alias_type == alias_type, Alias.alias_type_id == alias_type_id)
        )
        return AllAliasesSchema(aliases=aliases)
    except Exception as e:
        return APIResponse(status=500, message=f"Internal server error: {str(e)}")


@alias_api.post(
    "/{alias_type}/{alias_type_id}/add",
    response_model=APIResponse,
    summary="添加别名",
    description="管理员直接添加，非管理员提交审核",
)
async def add_alias(
    alias_type: AliasType,
    alias_type_id: int,
    data: AliasSchema = Depends(parse_json_body(engine, AliasSchema)),
    _: None = Depends(verify_api_auth),
) -> APIResponse:
    if not await is_alias_admin(engine, data.im_id):
        await engine.add_data(
            PendingAlias(
                alias_type=alias_type,
                alias_type_id=alias_type_id,
                alias=data.alias,
                submitted_by=data.im_id,
                submitted_at=datetime.now(UTC),
            )
        )
        return APIResponse(message="Alias submitted for review.")

    await engine.add_data(Alias(alias_type=alias_type, alias_type_id=alias_type_id, alias=data.alias))
    await redis_client.delete(f"pjsk-aliases:{alias_type}:{alias_type_id}")
    return APIResponse(message="Alias added.")


@alias_api.delete(
    "/{alias_type}/{alias_type_id}/{internal_id}",
    response_model=APIResponse,
    summary="删除别名",
    description="仅管理员可删除别名",
)
async def remove_alias(
    alias_type: AliasType,
    alias_type_id: int,
    internal_id: int,
    data: AliasSchema = Depends(parse_json_body(engine, AliasSchema, _require_admin=True)),
    _: None = Depends(verify_api_auth),
) -> APIResponse:
    await engine.delete_data(
        Alias,
        and_(
            Alias.alias_type == alias_type,
            Alias.alias_type_id == alias_type_id,
            Alias.alias == data.alias,
            Alias.id == internal_id,
        ),
    )
    await redis_client.delete(f"pjsk-aliases:{alias_type}:{alias_type_id}")
    return APIResponse(message="Alias deleted")


@alias_api.get(
    "/pending",
    response_model=Union[PendingAliasList, APIResponse],
    summary="获取待审核别名",
    description="管理员获取所有待审核别名",
)
async def get_pending_alias(
    _: None = Depends(require_alias_admin),
    __: None = Depends(verify_api_auth),
) -> Union[PendingAliasList, APIResponse]:
    rows = await engine.select_data(PendingAlias)
    if not rows:
        return APIResponse(status=404, message="No pending review alias")
    return PendingAliasList(rows=len(rows), results=[PendingAliasEntry.model_validate(row) for row in rows])


@alias_api.post(
    "/pending/{pending_id}/approve",
    response_model=APIResponse,
    summary="审核通过别名申请",
    description="管理员审核通过待审核别名申请",
)
async def approve_alias(
    pending_id: str,
    _: None = Depends(parse_json_body(engine, AliasApprovalSchema, _require_admin=True)),
    __: None = Depends(verify_api_auth),
) -> APIResponse:
    row = await engine.select_data(PendingAlias, PendingAlias.id == int(pending_id), one_result=True)
    if not row:
        return APIResponse(status=404, message="Pending alias not found")
    await engine.add_data(Alias(alias_type=row.alias_type, alias_type_id=row.alias_type_id, alias=row.alias))
    await engine.delete_data(PendingAlias, PendingAlias.id == int(pending_id))
    await redis_client.delete(f"pjsk-aliases:{row.alias_type}:{row.alias_type_id}")
    return APIResponse(message="Alias approved and added.")


@alias_api.post(
    "/pending/{pending_id}/reject",
    response_model=APIResponse,
    summary="审核拒绝别名申请",
    description="管理员拒绝待审核别名并记录理由申请",
)
async def reject_alias(
    pending_id: str,
    data: AliasRejectionSchema = Depends(parse_json_body(engine, AliasRejectionSchema, _require_admin=True)),
    _: None = Depends(verify_api_auth),
) -> APIResponse:
    row = await engine.select_data(PendingAlias, PendingAlias.id == int(pending_id), one_result=True)
    if not row:
        return APIResponse(status=404, message="Pending alias not found")
    await engine.add_data(
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
    await engine.delete_data(PendingAlias, PendingAlias.id == int(pending_id))
    return APIResponse(message="Alias rejected and logged.")


@alias_api.get(
    "/status/{pending_id}",
    response_model=APIResponse,
    summary="获取别名审核状态",
    description="查询别名审核状态和拒绝理由",
)
async def get_alias_review_status(
    pending_id: str,
    _: None = Depends(verify_api_auth),
) -> APIResponse:
    pending = await engine.select_data(PendingAlias, PendingAlias.id == int(pending_id), one_result=True)
    if pending:
        return APIResponse(status=202, message="This alias is pending review.")
    reason: Optional[str] = await engine.select_data(
        RejectedAlias.reason, RejectedAlias.id == int(pending_id), one_result=True
    )
    if reason:
        return APIResponse(status=400, message=reason)
    return APIResponse(status=404, message="Alias review record not found.")


@alias_api.get(
    "/group/{group_id}/{alias_type}/{alias_type_id}",
    response_model=Union[AllAliasesSchema, APIResponse],
    summary="获取群组别名",
    description="获取指定群组对应的别名列表",
)
async def get_group_alias(
    group_id: str, alias_type: AliasType, alias_type_id: int
) -> Union[AllAliasesSchema, APIResponse]:
    try:
        cache_key = f"pjsk-group-aliases:{group_id}:{alias_type}:{alias_type_id}"
        aliases = await get_cached_data(
            engine,
            cache_key,
            GroupAlias.alias,
            and_(
                GroupAlias.group_id == group_id,
                GroupAlias.alias_type == alias_type,
                GroupAlias.alias_type_id == alias_type_id,
            ),
        )
        if not aliases:
            return APIResponse(status=404, message="No aliases found for this group")
        return AllAliasesSchema(aliases=aliases)
    except Exception as e:
        return APIResponse(status=500, message=f"Internal server error: {str(e)}")


@alias_api.post(
    "/group/{group_id}/{alias_type}/{alias_type_id}",
    response_model=APIResponse,
    summary="添加群组别名",
    description="为指定群组添加别名",
)
async def add_group_alias(
    group_id: str,
    alias_type: AliasType,
    alias_type_id: int,
    data: AliasSchema = Depends(parse_json_body(engine, AliasSchema)),
    _: None = Depends(verify_api_auth),
) -> APIResponse:
    await engine.add_data(
        GroupAlias(group_id=group_id, alias_type=alias_type, alias_type_id=alias_type_id, alias=data.alias)
    )
    await redis_client.delete(f"pjsk-group-aliases:{group_id}:{alias_type}:{alias_type_id}")
    return APIResponse(message="Group alias added")


@alias_api.delete(
    "/group/{group_id}/{alias_type}/{alias_type_id}",
    response_model=APIResponse,
    summary="删除群组别名",
    description="删除指定群组的别名",
)
async def remove_group_alias(
    group_id: str,
    alias_type: AliasType,
    alias_type_id: int,
    data: AliasSchema = Depends(parse_json_body(engine, AliasSchema)),
    _: None = Depends(verify_api_auth),
) -> APIResponse:
    await engine.delete_data(
        GroupAlias,
        and_(
            GroupAlias.group_id == group_id,
            GroupAlias.alias_type == alias_type,
            GroupAlias.alias_type_id == alias_type_id,
            GroupAlias.alias == data.alias,
        ),
    )
    await redis_client.delete(f"pjsk-group-aliases:{group_id}:{alias_type}:{alias_type_id}")
    return APIResponse(message="Group alias deleted")
