from fastapi import APIRouter

from .alias import alias_api
from .binding import binding_api
from .music import music_api
chunithm_api = APIRouter(prefix="/chunithm", tags=["chunithm"])
chunithm_api.include_router(alias_api)
chunithm_api.include_router(binding_api)
chunithm_api.include_router(music_api)