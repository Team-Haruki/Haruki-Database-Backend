from quart import Blueprint, request
from pydantic import ValidationError
from sqlalchemy import select, delete, and_

from api.utils import success, error
from modules.sql.tables.pjsk import MusicAliases, GroupMusicAliases

from ..db_engine import engine
from ..schema import AliasBodySchema

music_alias_api = Blueprint("music_alias", __name__, url_prefix="/music")


@music_alias_api.route("/by_alias", methods=["GET"])
async def get_music_id_by_alias():
    alias = request.args.get("alias")
    group_id = request.args.get("group_id")
    if not alias:
        return error("Missing alias")

    async with engine.session() as session:
        if group_id:
            stmt = select(GroupMusicAliases.music_id).where(
                and_(GroupMusicAliases.alias == alias, GroupMusicAliases.group_id == group_id)
            )
        else:
            stmt = select(MusicAliases.music_id).where(MusicAliases.alias == alias)

        result = await session.execute(stmt)
        row = result.scalar_one_or_none()
        if row is None:
            return error("Alias not found")
        return success(row)


@music_alias_api.route("/<int:music_id>/all_aliases", methods=["GET"])
async def get_aliases_by_music_id(music_id):
    group_id = request.args.get("group_id")
    async with engine.session() as session:
        aliases = set()

        result = await session.execute(select(MusicAliases.alias).where(MusicAliases.music_id == music_id))
        aliases.update([row[0] for row in result.fetchall()])

        if group_id:
            result = await session.execute(
                select(GroupMusicAliases.alias).where(
                    and_(GroupMusicAliases.music_id == music_id, GroupMusicAliases.group_id == group_id)
                )
            )
            aliases.update([row[0] for row in result.fetchall()])

        return success(sorted(list(aliases)))


@music_alias_api.route("/<int:music_id>/alias", methods=["POST"])
async def add_alias(music_id):
    try:
        data = AliasBodySchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())

    async with engine.session() as session:
        if data.group_id:
            session.add(GroupMusicAliases(music_id=music_id, alias=data.alias, group_id=data.group_id))
        else:
            session.add(MusicAliases(music_id=music_id, alias=data.alias))
        await session.commit()
        return success(message="Alias added")


@music_alias_api.route("/<int:music_id>/alias", methods=["DELETE"])
async def delete_alias(music_id):
    try:
        data = AliasBodySchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())

    async with engine.session() as session:
        if data.group_id:
            await session.execute(
                delete(GroupMusicAliases).where(
                    and_(
                        GroupMusicAliases.music_id == music_id,
                        GroupMusicAliases.alias == data.alias,
                        GroupMusicAliases.group_id == data.group_id
                    )
                )
            )
        else:
            await session.execute(
                delete(MusicAliases).where(
                    and_(
                        MusicAliases.music_id == music_id,
                        MusicAliases.alias == data.alias
                    )
                )
            )
        await session.commit()
        return success(message="Alias deleted")
