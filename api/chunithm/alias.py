from fastapi_limiter.depends import RateLimiter
from fastapi_cache import FastAPICache
from sqlalchemy import and_
from fastapi import APIRouter, Query, Depends
from fastapi_cache.decorator import cache

from modules.exceptions import APIException
from modules.schemas.chunithm import MusicAliasSchema, AliasToMusicIDResponse, AllAliasesResponse, AddAliasResponse
from modules.schemas.response import APIResponse
from utils import chunithm_music_engine as engine, parse_json_body, verify_api_auth
from modules.sql.tables.chunithm import ChunithmMusicAlias

alias_api = APIRouter(prefix="/alias", tags=["alias_api"])


@alias_api.get(
    "/music-id",
    response_model=AliasToMusicIDResponse,
    summary="别名反查乐曲ID",
    description="通过提供的别名返回符合该别名的所有歌曲ID",
    dependencies=[Depends(RateLimiter(times=3, seconds=1))],
)
@cache(expire=300)
async def get_music_id(alias: str = Query(..., description="需要查询的别名")) -> AliasToMusicIDResponse:
    music_ids = await engine.select(ChunithmMusicAlias.music_id, ChunithmMusicAlias.alias == alias)
    if not music_ids:
        raise APIException(status=404, message="Alias not found")
    return AliasToMusicIDResponse(data=music_ids)


@alias_api.get(
    "/{music_id}",
    response_model=AllAliasesResponse,
    summary="获取乐曲别名",
    description="通过提供的乐曲ID返回该乐曲所有别名",
    dependencies=[Depends(RateLimiter(times=3, seconds=1))],
)
@cache(expire=300)
async def get_music_alias(music_id: int) -> AllAliasesResponse:
    aliases = await engine.select(ChunithmMusicAlias.alias, ChunithmMusicAlias.music_id == music_id)
    return AllAliasesResponse(data=aliases)


@alias_api.post(
    "/{music_id}/add",
    response_model=AllAliasesResponse,
    summary="添加乐曲别名",
    description="向提供的乐曲ID添加别名",
    dependencies=[Depends(verify_api_auth)],
)
async def add_alias(
    music_id: int, data: MusicAliasSchema = Depends(parse_json_body(engine, MusicAliasSchema))
) -> AddAliasResponse:
    new_alias = await engine.add(ChunithmMusicAlias(music_id=music_id, alias=data.alias))
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/alias/music-id?alias={data.alias}")
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/alias/{music_id}")
    return AddAliasResponse(message="Alias added", data=MusicAliasSchema(id=new_alias.id, alias=new_alias.alias))


@alias_api.delete(
    "/{music_id}/{internal_id}",
    response_model=APIResponse,
    summary="删除乐曲别名",
    description="删除提供的乐曲ID的特定别名",
    dependencies=[Depends(verify_api_auth)],
)
async def remove_alias(music_id: int, internal_id: int):
    result = await engine.delete(
        ChunithmMusicAlias, and_(ChunithmMusicAlias.music_id == music_id, ChunithmMusicAlias.id == internal_id)
    )
    if result == 0:
        raise APIException(status=404, message="Alias not found")
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/alias/{music_id}")
    return APIResponse(message="Alias deleted")
