from sqlalchemy import and_
from fastapi_cache import FastAPICache
from fastapi_cache.decorator import cache
from fastapi import APIRouter, Query, Depends
from fastapi_limiter.depends import RateLimiter

from modules.exceptions import APIException
from modules.schemas.response import APIResponse
from modules.cache_helpers import ORJsonCoder, cache_key_builder, clear_cache_by_path
from modules.schemas.chunithm import MusicAliasSchema, AliasToMusicIDResponse, AllAliasesResponse, AddAliasResponse
from utils import chunithm_binding_engine as engine, parse_json_body, verify_api_auth
from modules.sql.tables.chunithm import ChunithmMusicAlias

alias_api = APIRouter(prefix="/alias", tags=["CHUNITHM-Alias-API"])


@alias_api.get(
    "/music-id",
    response_model=AliasToMusicIDResponse,
    summary="别名反查乐曲ID",
    description="通过提供的别名返回符合该别名的所有歌曲ID",
    dependencies=[Depends(RateLimiter(times=3, seconds=1))],
)
@cache(expire=300, namespace="chunithm_alias", coder=ORJsonCoder, key_builder=cache_key_builder)  # type: ignore
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
@cache(expire=300, namespace="chunithm_alias", coder=ORJsonCoder, key_builder=cache_key_builder)  # type: ignore
async def get_music_alias(music_id: int) -> AllAliasesResponse:
    aliases = await engine.select(ChunithmMusicAlias.alias, ChunithmMusicAlias.music_id == music_id)
    return AllAliasesResponse(data=aliases)


@alias_api.post(
    "/{music_id}/add",
    response_model=AddAliasResponse,
    summary="添加乐曲别名",
    description="向提供的乐曲ID添加别名",
    dependencies=[Depends(verify_api_auth)],
)
async def add_alias(
    music_id: int, data: MusicAliasSchema = Depends(parse_json_body(engine, MusicAliasSchema))
) -> AddAliasResponse:
    existing = await engine.select(
        ChunithmMusicAlias, and_(ChunithmMusicAlias.music_id == music_id, ChunithmMusicAlias.alias == data.alias)
    )
    if existing:
        raise APIException(status=409, message="Alias already exists")
    new_alias = await engine.add(ChunithmMusicAlias(music_id=music_id, alias=data.alias))
    await clear_cache_by_path(namespace="chunithm_alias", path=f"/chunithm/alias/{music_id}")
    await clear_cache_by_path(
        namespace="chunithm_alias", path="/chunithm/alias/music-id", query_string=f"alias={data.alias}"
    )
    return AddAliasResponse(message="Alias added", data=MusicAliasSchema(id=new_alias.id, alias=new_alias.alias))


@alias_api.delete(
    "/{music_id}",
    response_model=APIResponse,
    summary="删除乐曲别名",
    description="删除提供的乐曲ID的特定别名",
    dependencies=[Depends(verify_api_auth)],
)
async def remove_alias(
    music_id: int, data: MusicAliasSchema = Depends(parse_json_body(engine, MusicAliasSchema))
) -> APIResponse:
    result = await engine.delete(
        ChunithmMusicAlias, and_(ChunithmMusicAlias.music_id == music_id, ChunithmMusicAlias.alias == data.alias)
    )
    if result == 0:
        raise APIException(status=404, message="Alias not found")
    await clear_cache_by_path(namespace="chunithm_alias", path=f"/chunithm/alias/{music_id}")
    await clear_cache_by_path(
        namespace="chunithm_alias", path="/chunithm/alias/music-id", query_string=f"alias={data.alias}"
    )
    return APIResponse(message="Alias deleted")
