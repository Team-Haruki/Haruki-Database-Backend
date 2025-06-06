from typing import Tuple
from pydantic import ValidationError
from sqlalchemy import select, delete
from quart import Blueprint, request, Response

from ..schema import MusicAliasSchema
from ..db_engine import music_engine as engine
from api.utils import success, error
from modules.sql.tables.chunithm import ChunithmMusicAlias

alias_api = Blueprint("alias_api", __name__, url_prefix="/alias")


@alias_api.route("/music-id", methods=["GET"])
async def get_music_id() -> Tuple[Response, int]:
    alias = request.args.get("alias")
    if alias is None:
        return error("Missing alias parameter", code=400)
    async with engine.session() as session:
        stmt = select(ChunithmMusicAlias.music_id).where(ChunithmMusicAlias.alias == alias)
        result = await session.execute(stmt)
        music_ids = result.scalars().all()
        if not music_ids:
            return error("Alias not found", code=404)
        return success({"music_ids": music_ids})


@alias_api.route("/<int:music_id>", methods=["GET"])
async def get_music_alias(music_id: int) -> Tuple[Response, int]:
    async with engine.session() as session:
        stmt = select(ChunithmMusicAlias.alias).where(ChunithmMusicAlias.music_id == music_id)
        result = await session.execute(stmt)
        aliases = result.scalars().all()
        return success({"aliases": aliases})


@alias_api.route("/<int:music_id>/add", methods=["POST"])
async def add_alias(music_id: int) -> Tuple[Response, int]:
    try:
        data = MusicAliasSchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())

    async with engine.session() as session:
        new_alias = ChunithmMusicAlias(music_id=music_id, alias=data.alias)
        session.add(new_alias)
        await session.commit()
        return success({"message": "Alias added", "id": new_alias.id})


@alias_api.route("/<int:music_id>/<int:internal_id>", methods=["DELETE"])
async def remove_alias(music_id: int, internal_id: int) -> Tuple[Response, int]:
    async with engine.session() as session:
        stmt = delete(ChunithmMusicAlias).where(
            ChunithmMusicAlias.music_id == music_id, ChunithmMusicAlias.id == internal_id
        )
        result = await session.execute(stmt)
        if result.rowcount == 0:
            return error("Alias not found", code=404)
        await session.commit()
        return success({"message": "Alias deleted"})
