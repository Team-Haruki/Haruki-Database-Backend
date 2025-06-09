from redis.asyncio import Redis
from fastapi import FastAPI, Request
from fastapi_cache import FastAPICache
from contextlib import asynccontextmanager
from fastapi.responses import ORJSONResponse
from fastapi_limiter import FastAPILimiter
from fastapi_cache.backends.redis import RedisBackend

from modules.exceptions import APIException
from modules.schemas.response import APIResponse
from configs.pjsk import PJSK_ENABLED
from configs.chunithm import CHUNITHM_ENABLED
from configs.redis import REDIS_HOST, REDIS_PORT, REDIS_PASSWORD


@asynccontextmanager
async def lifespan(_app: FastAPI):
    if PJSK_ENABLED:
        from utils import pjsk_engine

        await pjsk_engine.init_engine()
    if CHUNITHM_ENABLED:
        from utils import chunithm_bind_engine, chunithm_music_engine

        await chunithm_bind_engine.init_engine()
        await chunithm_music_engine.init_engine()
    redis_client = Redis(
        host=REDIS_HOST, port=REDIS_PORT, password=REDIS_PASSWORD, decode_responses=False, encoding="utf-8"
    )
    FastAPICache.init(RedisBackend(redis_client), prefix="fastapi-cache")
    await FastAPILimiter.init(redis_client)
    yield
    if PJSK_ENABLED:
        from utils import pjsk_engine

        await pjsk_engine.shutdown_engine()
    if CHUNITHM_ENABLED:
        from utils import chunithm_bind_engine, chunithm_music_engine

        await chunithm_bind_engine.shutdown_engine()
        await chunithm_music_engine.shutdown_engine()
    await FastAPILimiter.close()


app = FastAPI(lifespan=lifespan, default_response_class=ORJSONResponse)
if PJSK_ENABLED:
    from api.pjsk.core import pjsk_api

    app.include_router(pjsk_api)
if CHUNITHM_ENABLED:
    from api.chunithm.core import chunithm_api

    app.include_router(chunithm_api)


@app.exception_handler(APIException)
async def api_exception_handler(request: Request, exc: APIException) -> ORJSONResponse:
    return ORJSONResponse(
        status_code=exc.status,
        content=APIResponse(
            status=exc.status,
            message=exc.message,
        ).model_dump(),
    )
