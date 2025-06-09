from sqlalchemy import and_
from fastapi import APIRouter, Depends
from fastapi_cache import FastAPICache
from fastapi_cache.decorator import cache

from modules.exceptions import APIException
from modules.schemas.response import APIResponse
from utils import chunithm_bind_engine as engine
from modules.schemas.chunithm import DefaultServerResponse, BindingResponse
from utils import verify_api_auth
from modules.sql.tables.chunithm import ChunithmBinding, ChunithmDefaultServer

binding_api = APIRouter(prefix="/{platform}/user", tags=["binding_api"])


@binding_api.get(
    "/{im_id}/default",
    response_model=DefaultServerResponse,
    summary="获取默认绑定服务器",
    description="获取用户默认绑定的chunithm服务器",
    dependencies=[Depends(verify_api_auth)],
)
@cache(expire=300)
async def get_default_server(platform: str, im_id: str) -> DefaultServerResponse:
    server = await engine.select(
        ChunithmDefaultServer.server,
        and_(ChunithmDefaultServer.platform == platform, ChunithmDefaultServer.im_id == im_id),
        one_result=True,
    )
    if server is None:
        raise APIException(status=404, message="Default server not set")
    return DefaultServerResponse.model_validate(server)


@binding_api.get(
    "/{im_id}/{server}",
    response_model=ChunithmBinding,
    summary="获取绑定信息",
    description="获取用户chunithm的绑定信息(hdd)",
    dependencies=[Depends(verify_api_auth)],
)
@cache(expire=300)
async def get_binding(platform: str, im_id: str, server: str) -> BindingResponse:
    binding = await engine.select(
        ChunithmBinding,
        and_(ChunithmBinding.platform == platform, ChunithmBinding.im_id == im_id, ChunithmBinding.server == server),
        one_result=True,
    )
    if binding is None:
        raise APIException(status=404, message="Binding not found")
    return BindingResponse.model_validate(binding)


@binding_api.put(
    "/{im_id}/{server}/{aime_id}",
    response_model=BindingResponse,
    summary="添加/更新绑定",
    description="添加/更新用户的chunithm绑定信息(hdd)",
    dependencies=[Depends(verify_api_auth)],
)
async def update_binding(platform: str, im_id: str, server: str, aime_id: str) -> APIResponse:
    existing = await engine.select(
        ChunithmBinding, and_(ChunithmBinding.im_id == im_id, ChunithmBinding.server == server), one_result=True
    )
    if existing:
        existing.aime_id = aime_id
    else:
        await engine.add(ChunithmBinding(platform=platform, im_id=im_id, server=server, aime_id=aime_id))
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/{server}")
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/default")
    return APIResponse(message="Binding updated")


@binding_api.delete(
    "/{im_id}/{server}/{aime_id}",
    response_model=APIResponse,
    summary="删除绑定",
    description="删除用户的chunithm绑定信息(hdd)",
    dependencies=[Depends(verify_api_auth)],
)
async def delete_binding(platform: str, im_id: str, server: str, aime_id: str) -> APIResponse:
    binding = await engine.select(
        ChunithmBinding,
        and_(
            ChunithmBinding.platform == platform,
            ChunithmBinding.im_id == im_id,
            ChunithmBinding.server == server,
            ChunithmBinding.aime_id == aime_id,
        ),
        one_result=True,
    )
    if binding is None:
        raise APIException(status=404, message="Binding not found")
    async with engine.session() as session:
        await session.delete(binding)
        await session.commit()
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/{server}")
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/default")
    return APIResponse(message="Binding deleted")
