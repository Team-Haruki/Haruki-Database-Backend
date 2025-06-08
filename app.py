from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from contextlib import asynccontextmanager

from modules.exceptions import APIException
from modules.schemas.response import APIResponse
from configs.pjsk import PJSK_ENABLED
from configs.chunithm import CHUNITHM_ENABLED


@asynccontextmanager
async def lifespan(_app: FastAPI):
    if PJSK_ENABLED:
        from utils import pjsk_engine

        await pjsk_engine.init_engine()
    if CHUNITHM_ENABLED:
        from utils import chunithm_bind_engine, chunithm_music_engine

        await chunithm_bind_engine.init_engine()
        await chunithm_music_engine.init_engine()
    yield
    if PJSK_ENABLED:
        from utils import pjsk_engine

        await pjsk_engine.shutdown_engine()
    if CHUNITHM_ENABLED:
        from utils import chunithm_bind_engine, chunithm_music_engine

        await chunithm_bind_engine.shutdown_engine()
        await chunithm_music_engine.shutdown_engine()


app = FastAPI(lifespan=lifespan)
if PJSK_ENABLED:
    from api.pjsk.core import pjsk_api

    app.include_router(pjsk_api)
if CHUNITHM_ENABLED:
    from api.chunithm.core import chunithm_api

    app.include_router(chunithm_api)


@app.exception_handler(APIException)
async def api_exception_handler(request: Request, exc: APIException) -> JSONResponse:
    return JSONResponse(
        status_code=exc.status,
        content=APIResponse(
            status=exc.status,
            message=exc.message,
        ).model_dump(),
    )
