import orjson
from fastapi import Request, Query
from pydantic import ValidationError
from sqlalchemy import select
from sqlalchemy.orm import InstrumentedAttribute
from typing import Optional, Type, Union, TypeVar

from modules.redis import RedisClient
from modules.exceptions import APIException
from modules.sql.engine import DatabaseEngine
from modules.sql.tables.base import Base
from modules.sql.tables.pjsk import AliasAdmin
from configs.pjsk import PJSK_DB_URL
from configs.app import ACCPET_AUTHORIZATION, ACCEPT_USER_AGENT
from configs.redis import REDIS_HOST, REDIS_PORT, REDIS_PASSWORD
from configs.chunithm import CHUNITHM_BIND_DB_URL, CHUNITHM_MUSIC_DB_URL

redis_client: Optional[RedisClient] = RedisClient(REDIS_HOST, REDIS_PORT, REDIS_PASSWORD)
chunithm_bind_engine: Optional[DatabaseEngine] = DatabaseEngine(CHUNITHM_BIND_DB_URL)
chunithm_music_engine: Optional[DatabaseEngine] = DatabaseEngine(CHUNITHM_MUSIC_DB_URL)
pjsk_engine: Optional[DatabaseEngine] = DatabaseEngine(PJSK_DB_URL)

T = TypeVar("T")


async def verify_api_auth(request: Request) -> None:
    auth_header = request.headers.get("Authorization")
    user_agent = request.headers.get("User-Agent")
    if auth_header != ACCPET_AUTHORIZATION:
        raise APIException(status=401, message="Invalid Authorization header")
    if ACCEPT_USER_AGENT and user_agent != ACCEPT_USER_AGENT:
        raise APIException(status=403, message="Invalid User-Agent")


async def is_alias_admin(engine: DatabaseEngine, im_id: str) -> bool:
    async with engine.session() as session:
        result = await session.execute(select(AliasAdmin).where(AliasAdmin.im_id == im_id))
        return result.scalar_one_or_none() is not None


async def require_alias_admin(engine: DatabaseEngine, im_id: str = Query(..., description="IM user ID")) -> None:
    if not await is_alias_admin(engine, im_id):
        raise APIException(status=401, message="Permission denied")


async def get_cached_data(
    engine: DatabaseEngine, key: str, target: Union[Type[Base], InstrumentedAttribute], *conditions
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
