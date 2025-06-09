from sqlalchemy import and_
from datetime import datetime, UTC
from typing import Optional, Union
from fastapi_cache import FastAPICache
from fastapi_cache.decorator import cache
from fastapi import APIRouter, Query, Depends

from modules.exceptions import APIException
from modules.enums import AliasType
from modules.schemas.pjsk import (
    AliasSchema,
    AllAliasesSchema,
    PendingAliasList,
    PendingAliasEntry,
    AliasApprovalSchema,
    AliasRejectionSchema,
    AliasToObjectIdSchema,
)
from modules.schemas.response import APIResponse
from modules.sql.tables.pjsk import (
    Alias,
    GroupAlias,
    PendingAlias,
    RejectedAlias,
)
from utils import is_alias_admin, require_alias_admin, parse_json_body, verify_api_auth
from utils import pjsk_engine as engine

alias_api = APIRouter(prefix="/alias", tags=["Alias"])


@alias_api.get(
    "/{alias_type}-id",
    response_model=AliasToObjectIdSchema,
    summary="根据别名获取目标类型ID",
    description="根据歌曲/角色别名返回所有对应类型ID",
)
@cache(expire=300)
async def get_alias_type_id(
    alias_type: AliasType,
    alias: str = Query(..., description="Alias to lookup"),
) -> AliasToObjectIdSchema:
    try:
        select_object, select_clause = (
            Alias.alias_type_id,
            and_(Alias.alias_type == alias_type, Alias.alias == alias),
        )
        rows = await engine.select(select_object, select_clause)
        if not rows:
            raise APIException(status=404, message="Alias not found")
        return AliasToObjectIdSchema(match_ids=rows)
    except Exception as e:
        raise APIException(status=500, message=f"Internal server error: {str(e)}")


@alias_api.get(
    "/{alias_type}/{alias_type_id}",
    response_model=AllAliasesSchema,
    summary="获取指定别名与指定别名类型ID的全部别名",
    description="根据歌曲/角色ID返回所有对应别名",
)
@cache(expire=300)
async def get_alias(alias_type: AliasType, alias_type_id: int) -> AllAliasesSchema:
    try:
        aliases = await engine.select(
            Alias.alias, and_(Alias.alias_type == alias_type, Alias.alias_type_id == alias_type_id)
        )
        return AllAliasesSchema(aliases=aliases)
    except Exception as e:
        raise APIException(status=500, message=f"Internal server error: {str(e)}")


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
        await engine.add(
            PendingAlias(
                alias_type=alias_type,
                alias_type_id=alias_type_id,
                alias=data.alias,
                submitted_by=data.im_id,
                submitted_at=datetime.now(UTC),
            )
        )
        return APIResponse(message="Alias submitted for review.")

    await engine.add(Alias(alias_type=alias_type, alias_type_id=alias_type_id, alias=data.alias))
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/pjsk/alias/{alias_type}-id?alias={data.alias}")
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/pjsk/alias/{alias_type}/{alias_type_id}")
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
    await engine.delete(
        Alias,
        and_(
            Alias.alias_type == alias_type,
            Alias.alias_type_id == alias_type_id,
            Alias.alias == data.alias,
            Alias.id == internal_id,
        ),
    )
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/pjsk/alias/{alias_type}-id?alias={data.alias}")
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/pjsk/alias/{alias_type}/{alias_type_id}")
    return APIResponse(message="Alias deleted")


@alias_api.get(
    "/pending",
    response_model=PendingAliasList,
    summary="获取待审核别名",
    description="管理员获取所有待审核别名",
)
async def get_pending_alias(
    _: None = Depends(require_alias_admin),
    __: None = Depends(verify_api_auth),
) -> Union[PendingAliasList, APIResponse]:
    rows = await engine.select(PendingAlias)
    if not rows:
        raise APIException(status=404, message="No pending review alias")
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
    row = await engine.select(PendingAlias, PendingAlias.id == int(pending_id), one_result=True)
    if not row:
        raise APIException(status=404, message="Pending alias not found")
    await engine.add(Alias(alias_type=row.alias_type, alias_type_id=row.alias_type_id, alias=row.alias))
    await engine.delete(PendingAlias, PendingAlias.id == int(pending_id))
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/pjsk/alias/{row.alias_type}-id?alias={row.alias}")
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/pjsk/alias/{row.alias_type}/{row.alias_type_id}")
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/pjsk/alias/status/{pending_id}")
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
    row = await engine.select(PendingAlias, PendingAlias.id == int(pending_id), one_result=True)
    if not row:
        return APIResponse(status=404, message="Pending alias not found")
    await engine.add(
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
    await engine.delete(PendingAlias, PendingAlias.id == int(pending_id))
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/pjsk/alias/status/{pending_id}")
    return APIResponse(message="Alias rejected and logged.")


@alias_api.get(
    "/status/{pending_id}",
    summary="获取别名审核状态",
    description="查询别名审核状态和拒绝理由",
)
@cache(expire=300)
async def get_alias_review_status(
    pending_id: str,
    _: None = Depends(verify_api_auth),
) -> APIResponse:
    pending = await engine.select(PendingAlias, PendingAlias.id == int(pending_id), one_result=True)
    if pending:
        raise APIException(status=202, message="This alias is pending review.")
    reason: Optional[str] = await engine.select(
        RejectedAlias.reason, RejectedAlias.id == int(pending_id), one_result=True
    )
    if reason:
        raise APIException(status=400, message=reason)
    raise APIException(status=404, message="Alias review record not found.")


@alias_api.get(
    "/group/{group_id}/{alias_type}/{alias_type_id}",
    response_model=AllAliasesSchema,
    summary="获取群组别名",
    description="获取指定群组对应的别名列表",
)
@cache(expire=300)
async def get_group_alias(
    group_id: str, alias_type: AliasType, alias_type_id: int
) -> Union[AllAliasesSchema, APIResponse]:
    try:
        aliases = await engine.select(
            GroupAlias.alias,
            and_(
                GroupAlias.group_id == group_id,
                GroupAlias.alias_type == alias_type,
                GroupAlias.alias_type_id == alias_type_id,
            ),
        )
        if not aliases:
            raise APIException(status=404, message="No aliases found for this group")
        return AllAliasesSchema(aliases=aliases)
    except Exception as e:
        raise APIException(status=500, message=f"Internal server error: {str(e)}")


@alias_api.get(
    "/group/{alias_type}-id",
    response_model=AliasToObjectIdSchema,
    summary="根据群组别名获取目标类型ID",
    description="根据群组内歌曲/角色别名返回所有对应类型ID",
)
@cache(expire=300)
async def get_group_alias_type_id(
    alias_type: AliasType,
    alias: str = Query(..., description="Alias to lookup"),
    group_id: str = Query(..., description="Group ID"),
) -> AliasToObjectIdSchema:
    try:
        rows = await engine.select(
            GroupAlias.alias_type_id,
            and_(
                GroupAlias.alias_type == alias_type,
                GroupAlias.alias == alias,
                GroupAlias.group_id == group_id,
            ),
        )
        if not rows:
            raise APIException(status=404, message="Alias not found")
        return AliasToObjectIdSchema(match_ids=rows)
    except Exception as e:
        raise APIException(status=500, message=f"Internal server error: {str(e)}")


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
    await engine.add(
        GroupAlias(group_id=group_id, alias_type=alias_type, alias_type_id=alias_type_id, alias=data.alias)
    )
    await FastAPICache.clear(
        namespace="fastapi-cache", key=f"/pjsk/alias/group/{group_id}/{alias_type}/{alias_type_id}"
    )
    await FastAPICache.clear(
        namespace="fastapi-cache", key=f"/pjsk/alias/group/{alias_type}-id?alias={data.alias}&group_id={group_id}"
    )
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
    await engine.delete(
        GroupAlias,
        and_(
            GroupAlias.group_id == group_id,
            GroupAlias.alias_type == alias_type,
            GroupAlias.alias_type_id == alias_type_id,
            GroupAlias.alias == data.alias,
        ),
    )
    await FastAPICache.clear(
        namespace="fastapi-cache", key=f"/pjsk/alias/group/{group_id}/{alias_type}/{alias_type_id}"
    )
    await FastAPICache.clear(
        namespace="fastapi-cache", key=f"/pjsk/alias/group/{alias_type}-id?alias={data.alias}&group_id={group_id}"
    )
    return APIResponse(message="Group alias deleted")
