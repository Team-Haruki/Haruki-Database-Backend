from typing import Tuple
from pydantic import ValidationError
from sqlalchemy import select, delete
from fastapi import APIRouter, Request, HTTPException
from fastapi.responses import ORJSONResponse
from fastapi import status

from modules.schemas.chunithm import MusicAliasSchema
from utils import chunithm_music_engine as engine
from utils import success, error, redis_client
from modules.sql.tables.chunithm import ChunithmMusicAlias

alias_api = APIRouter(prefix="/alias", tags=["alias_api"])


@alias_api.get("/music-id")
async def get_music_id(request: Request):
    alias = request.query_params.get("alias")
    if alias is None:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Missing alias parameter")
    async with engine.session() as session:
        stmt = select(ChunithmMusicAlias.music_id).where(ChunithmMusicAlias.alias == alias)
        result = await session.execute(stmt)
        music_ids = result.scalars().all()
        if not music_ids:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Alias not found")
        return ORJSONResponse(content=success({"music_ids": music_ids}))


@alias_api.get("/{music_id}")
async def get_music_alias(music_id: int):
    cache_key = f"chunithm_aliases:{music_id}"
    cached = await redis_client.get(cache_key)
    if cached:
        return ORJSONResponse(content=success({"aliases": cached}))

    async with engine.session() as session:
        stmt = select(ChunithmMusicAlias.alias).where(ChunithmMusicAlias.music_id == music_id)
        result = await session.execute(stmt)
        aliases = result.scalars().all()
        await redis_client.set(cache_key, aliases)
        return ORJSONResponse(content=success({"aliases": aliases}))


@alias_api.post("/{music_id}/add")
async def add_alias(music_id: int, request: Request):
    try:
        data = MusicAliasSchema(**await request.json())
    except ValidationError as ve:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_ENTITY, detail=ve.errors())

    async with engine.session() as session:
        new_alias = ChunithmMusicAlias(music_id=music_id, alias=data.alias)
        session.add(new_alias)
        await session.commit()
        await redis_client.delete(f"chunithm_aliases:{music_id}")
        return ORJSONResponse(content=success({"id": new_alias.id}, message="Alias added"))


@alias_api.delete("/{music_id}/{internal_id}")
async def remove_alias(music_id: int, internal_id: int):
    async with engine.session() as session:
        stmt = delete(ChunithmMusicAlias).where(
            ChunithmMusicAlias.music_id == music_id, ChunithmMusicAlias.id == internal_id
        )
        result = await session.execute(stmt)
        if result.rowcount == 0:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Alias not found")
        await session.commit()
        await redis_client.delete(f"chunithm_aliases:{music_id}")
        return ORJSONResponse(content=success(message="Alias deleted"))
