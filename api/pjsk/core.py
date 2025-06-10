from fastapi import APIRouter

from .alias import alias_api
from .binding import binding_api
from .preference import preference_api

pjsk_api = APIRouter(prefix="/pjsk", tags=["pjsk"])
pjsk_api.include_router(alias_api)
pjsk_api.include_router(binding_api)
pjsk_api.include_router(preference_api)
