from quart import Blueprint, request
from pydantic import ValidationError
from sqlalchemy import select, delete, and_

from api.utils import success, error
from modules.sql.tables.pjsk import CharacterAliases, GroupCharacterAliases
from ..db_engine import engine
from ..schema import AliasBodySchema

character_alias_api = Blueprint("character_alias", __name__, url_prefix="/character")


@character_alias_api.route("/by_alias", methods=["GET"])
async def get_character_id_by_alias():
    alias = request.args.get("alias")
    group_id = request.args.get("group_id")
    if not alias:
        return error("Missing alias")

    async with engine.session() as session:
        if group_id:
            stmt = select(GroupCharacterAliases.character_id).where(
                and_(GroupCharacterAliases.alias == alias, GroupCharacterAliases.group_id == group_id)
            )
        else:
            stmt = select(CharacterAliases.character_id).where(CharacterAliases.alias == alias)

        result = await session.execute(stmt)
        row = result.scalar_one_or_none()
        if row is None:
            return error("Alias not found")
        return success(row)


@character_alias_api.route("/<int:character_id>/all_aliases", methods=["GET"])
async def get_aliases_by_character_id(character_id):
    group_id = request.args.get("group_id")
    async with engine.session() as session:
        aliases = set()

        result = await session.execute(
            select(CharacterAliases.alias).where(CharacterAliases.character_id == character_id))
        aliases.update([row[0] for row in result.fetchall()])

        if group_id:
            result = await session.execute(
                select(GroupCharacterAliases.alias).where(
                    and_(GroupCharacterAliases.character_id == character_id, GroupCharacterAliases.group_id == group_id)
                )
            )
            aliases.update([row[0] for row in result.fetchall()])

        return success(sorted(list(aliases)))


@character_alias_api.route("/<int:character_id>/alias", methods=["POST"])
async def add_alias(character_id):
    try:
        data = AliasBodySchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())

    async with engine.session() as session:
        if data.group_id:
            session.add(GroupCharacterAliases(character_id=character_id, alias=data.alias, group_id=data.group_id))
        else:
            session.add(CharacterAliases(character_id=character_id, alias=data.alias))
        await session.commit()
        return success(message="Alias added")


@character_alias_api.route("/<int:character_id>/alias", methods=["DELETE"])
async def delete_alias(character_id):
    try:
        data = AliasBodySchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())

    async with engine.session() as session:
        if data.group_id:
            await session.execute(
                delete(GroupCharacterAliases).where(
                    and_(
                        GroupCharacterAliases.character_id == character_id,
                        GroupCharacterAliases.alias == data.alias,
                        GroupCharacterAliases.group_id == data.group_id
                    )
                )
            )
        else:
            await session.execute(
                delete(CharacterAliases).where(
                    and_(
                        CharacterAliases.character_id == character_id,
                        CharacterAliases.alias == data.alias
                    )
                )
            )
        await session.commit()
        return success(message="Alias deleted")
