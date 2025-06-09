from typing import Optional
from sqlalchemy import update, and_
from fastapi_cache import FastAPICache
from fastapi_cache.decorator import cache
from fastapi import APIRouter, Query, Depends
from starlette.requests import Request

from modules.exceptions import APIException
from modules.schemas.response import APIResponse
from modules.enums import BindingServer, DefaultBindingServer
from modules.sql.tables.pjsk import UserBinding, UserDefaultBinding
from modules.schemas.pjsk import BindingSchema, BindingResponse, EditBindingSchema, AddBindingSuccessResponse
from utils import pjsk_engine as engine, parse_json_body, verify_api_auth

binding_api = APIRouter(prefix="/{platform}/user", tags=["user_binding"])


@binding_api.get(
    "/{im_id}/binding",
    response_model=BindingResponse,
    summary="获取绑定信息",
    description="根据平台和用户IM ID获取所有服务器绑定信息，可选筛选特定服务器",
    dependencies=[Depends(verify_api_auth)],
)
@cache(expire=300)
async def get_bindings(
    platform: str,
    im_id: str,
    server: Optional[BindingServer] = Query(None, description="Server code such as jp, cn, etc."),
) -> BindingResponse:
    clause = (
        and_(UserBinding.platform == platform, UserBinding.im_id == im_id, UserBinding.server == server)
        if server
        else and_(UserBinding.platform == platform, UserBinding.im_id == im_id)
    )
    results = await engine.select(UserBinding, clause, unique=True)
    if results:
        return BindingResponse(bindings=[BindingSchema(**binding) for binding in results])
    else:
        raise APIException(status=404, message="No bindings found")


@binding_api.post(
    "/{im_id}/binding",
    response_model=AddBindingSuccessResponse,
    status_code=201,
    summary="添加绑定信息",
    description="添加新的用户绑定记录，如果记录已存在则返回冲突错误",
    dependencies=[Depends(verify_api_auth)],
)
async def add_binding(
    platform: str,
    im_id: str,
    data: EditBindingSchema = Depends(parse_json_body(engine, EditBindingSchema)),
) -> AddBindingSuccessResponse:
    if data.server == DefaultBindingServer.default:
        raise APIException(status=400, message="Unacceptable server param")
    exists_result = await engine.select(
        UserBinding,
        and_(
            UserBinding.platform == platform,
            UserBinding.im_id == im_id,
            UserBinding.server == data.server,
            UserBinding.user_id == data.user_id,
        ),
        unique=True,
        one_result=True,
    )
    if exists_result:
        raise APIException(status=409, message="Binding already exists.")
    add_result = await engine.add(
        UserBinding(platform=platform, im_id=im_id, server=str(data.server), user_id=data.user_id, visible=data.visible)
    )
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/binding")
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/binding?server={data.server}")
    return AddBindingSuccessResponse(bind_id=add_result.id, status=201)


@binding_api.get(
    "/{im_id}/binding/default",
    response_model=BindingResponse,
    summary="获取默认绑定",
    description="获取某个用户在指定服务器或全局的默认绑定信息",
    dependencies=[Depends(verify_api_auth)],
)
@cache(expire=300)
async def get_default_binding(
    platform: str,
    im_id: str,
    server: Optional[DefaultBindingServer] = Query(
        DefaultBindingServer.default, description="Server code such as jp, cn, etc."
    ),
) -> BindingResponse:
    binding = await engine.select_with_join(
        UserBinding,
        UserDefaultBinding,
        and_(UserBinding.platform == platform, UserBinding.im_id == im_id, UserBinding.server == str(server)),
        unique=True,
        one_result=True,
    )
    if not binding:
        msg = f"No default for server '{server}'" if server != "default" else "No global default set"
        raise APIException(status=404, message=msg)
    return BindingResponse(binding=BindingSchema.model_validate(binding))


@binding_api.put(
    "/{im_id}/binding/default",
    response_model=APIResponse,
    summary="设置默认绑定",
    description="设置指定服务器（或全局）的默认绑定，替换原有绑定",
    dependencies=[Depends(verify_api_auth)],
)
async def set_default(
    platform: str,
    im_id: str,
    data: EditBindingSchema = Depends(parse_json_body(engine, EditBindingSchema)),
) -> APIResponse:
    binding = await engine.select(
        UserBinding, and_(UserBinding.platform == platform, UserBinding.im_id == im_id), unique=True, one_result=True
    )
    if not binding:
        raise APIException(status=404, message="Binding not found")
    if data.server != DefaultBindingServer.default:
        if str(data.server) != str(binding.server):
            raise APIException(status=400, message="Illegal request")
    await engine.delete(
        UserDefaultBinding,
        and_(
            UserDefaultBinding.platform == platform,
            UserDefaultBinding.im_id == im_id,
            UserDefaultBinding.server == str(data.server),
        ),
    )
    await engine.add(UserDefaultBinding(platform=platform, im_id=im_id, server=str(data.server), bind_id=data.bind_id))
    await FastAPICache.clear(
        namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/binding/default?server={data.server}"
    )
    return APIResponse(status=200, message=f"Set default binding for {data.server}")


@binding_api.delete(
    "/{im_id}/binding/default",
    response_model=APIResponse,
    summary="删除默认绑定",
    description="删除指定平台和服务器上的用户默认绑定记录",
    dependencies=[Depends(verify_api_auth)],
)
async def delete_default(
    platform: str,
    im_id: str,
    data: EditBindingSchema = Depends(parse_json_body(engine, EditBindingSchema)),
) -> APIResponse:
    await engine.delete(
        UserDefaultBinding,
        and_(
            UserDefaultBinding.platform == platform,
            UserDefaultBinding.im_id == im_id,
            UserDefaultBinding.server == data.server,
        ),
    )
    await FastAPICache.clear(
        namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/binding/default?server={data.server}"
    )
    return APIResponse(status=200, message=f"Deleted default binding for {data.server}")


@binding_api.patch(
    "/{im_id}/binding/{bind_id}",
    response_model=APIResponse,
    summary="更新绑定UID可见性",
    description="更改指定绑定项的UID可见性状态",
    dependencies=[Depends(verify_api_auth)],
)
async def update_visibility(
    platform: str,
    im_id: str,
    bind_id: int,
    data: EditBindingSchema = Depends(parse_json_body(engine, EditBindingSchema)),
) -> APIResponse:
    binding = await engine.select(
        UserBinding,
        and_(UserBinding.platform == platform, UserBinding.im_id == im_id, UserBinding.id == bind_id),
        unique=True,
        one_result=True,
    )
    if not binding:
        raise APIException(status=404, message="Binding not found")
    async with engine.session() as session:
        await session.execute(
            update(UserBinding)
            .where(UserBinding.platform == platform, UserBinding.id == bind_id)
            .values(visible=data.visible)
        )
        await session.commit()
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/binding")
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/binding?server={binding.server}")
    return APIResponse(status=200, message="Visibility updated")


@binding_api.delete(
    "/{im_id}/binding/{bind_id}",
    response_model=APIResponse,
    summary="删除绑定",
    description="彻底删除某条用户绑定记录及其默认绑定引用",
    dependencies=[Depends(verify_api_auth)],
)
async def delete_binding(platform: str, im_id: str, bind_id: int) -> APIResponse:
    binding = await engine.select(
        UserBinding,
        and_(UserBinding.platform == platform, UserBinding.im_id == im_id, UserBinding.id == bind_id),
        unique=True,
        one_result=True,
    )
    await engine.delete(
        UserDefaultBinding,
        and_(
            UserDefaultBinding.platform == platform,
            UserDefaultBinding.im_id == im_id,
            UserDefaultBinding.bind_id == bind_id,
        ),
    )
    await engine.delete(
        UserBinding, and_(UserBinding.platform == platform, UserBinding.id == bind_id, UserBinding.im_id == im_id)
    )
    if binding:
        await FastAPICache.clear(namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/binding")
        await FastAPICache.clear(
            namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/binding?server={binding.server}"
        )
        await FastAPICache.clear(
            namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/binding/default?server=default"
        )
        await FastAPICache.clear(
            namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/binding/default?server={binding.server}"
        )
    return APIResponse(status=200, message="Binding deleted")
