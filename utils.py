import orjson
from sqlalchemy import select
from fastapi import Request, Query
from pydantic import ValidationError
from sqlalchemy.orm import InstrumentedAttribute
from typing import Optional, Type, Union, TypeVar

from modules.exceptions import APIException
from modules.sql.engine import DatabaseEngine
from modules.sql.tables.pjsk import AliasAdmin
from modules.sql.tables.base import PjskBase, ChunithmMainBase, ChunithmMusicDBBase
from configs.pjsk import PJSK_DB_URL
from configs.app import ACCPET_AUTHORIZATION, ACCEPT_USER_AGENT
from configs.chunithm import CHUNITHM_BINDING_DB_URL, CHUNITHM_MUSIC_DB_URL

chunithm_bind_engine: Optional[DatabaseEngine] = DatabaseEngine(CHUNITHM_BINDING_DB_URL, table_base=ChunithmMusicDBBase)
chunithm_music_engine: Optional[DatabaseEngine] = DatabaseEngine(CHUNITHM_MUSIC_DB_URL, table_base=ChunithmMainBase)
pjsk_engine: Optional[DatabaseEngine] = DatabaseEngine(PJSK_DB_URL, table_base=PjskBase)

T = TypeVar("T")


async def verify_api_auth(request: Request) -> None:
    auth_header = request.headers.get("Authorization")
    user_agent = request.headers.get("User-Agent")
    if ACCPET_AUTHORIZATION and auth_header != ACCPET_AUTHORIZATION:
        raise APIException(status=401, message="Invalid Authorization header")
    if ACCEPT_USER_AGENT and ACCEPT_USER_AGENT not in user_agent:
        raise APIException(status=403, message="Invalid User-Agent")


async def is_alias_admin(engine: DatabaseEngine, im_id: str) -> bool:
    async with engine.session() as session:
        result = await session.execute(select(AliasAdmin).where(AliasAdmin.im_id == im_id))
        return result.scalar_one_or_none() is not None


def require_alias_admin(engine: DatabaseEngine):
    async def _require(im_id: str = Query(..., description="IM user ID")) -> None:
        if not await is_alias_admin(engine, im_id):
            raise APIException(status=401, message="Permission denied")
    return _require


async def get_cached_data(
    engine: DatabaseEngine, key: str, target: Union[Type[T], InstrumentedAttribute], *conditions
) -> T:
    cached = await redis_client.get(key)
    if cached:
        return orjson.loads(cached)
    aliases = await engine.select(target, *conditions)
    await redis_client.set(key, orjson.dumps(aliases))
    return aliases


def parse_json_body(engine: DatabaseEngine, model: Type[T], _require_admin: bool = False):
    async def dependency(request: Request) -> T:
        try:
            body = await request.json()
            obj = model(**body)
        except ValidationError as ve:
            raise APIException(status=422, message=f"Validation error: {ve.errors()}")
        if _require_admin and not await is_alias_admin(engine, getattr(obj, "im_id", None)):
            raise APIException(status=401, message="Permission denied")
        return obj

    return dependency
