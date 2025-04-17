from quart import Blueprint, request
from pydantic import ValidationError
from sqlalchemy import select, update, delete

from api.utils import success, error
from modules.sql.tables.pjsk import UserPreferences

from ..schema import UserPreferenceSchema
from ..db_engine import engine

preference_api = Blueprint("user_preference", __name__, url_prefix="/user")


@preference_api.route("/<int:im_id>/preference", methods=["GET"])
async def get_preferences(im_id):
    async with engine.session() as session:
        result = await session.execute(select(UserPreferences).where(UserPreferences.user_id == im_id))
        prefs = result.scalars().all()
        return success([{"option": p.option, "value": p.value} for p in prefs])


@preference_api.route("/<int:im_id>/preference/<option>", methods=["GET"])
async def get_preference_option(im_id, option):
    async with engine.session() as session:
        stmt = select(UserPreferences).where(UserPreferences.user_id == im_id, UserPreferences.option == option)
        result = await session.execute(stmt)
        pref = result.scalar_one_or_none()
        if pref:
            return success({"option": pref.option, "value": pref.value})
        return error("Preference not found", code=404)


@preference_api.route("/<int:im_id>/preference", methods=["POST"])
async def set_preference(im_id):
    try:
        data = UserPreferenceSchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())

    async with engine.session() as session:
        stmt = update(UserPreferences).where(
            UserPreferences.user_id == im_id,
            UserPreferences.option == data.option
        ).values(value=data.value)
        result = await session.execute(stmt)
        if result.rowcount == 0:
            session.add(UserPreferences(user_id=im_id, option=data.option, value=data.value))
        await session.commit()
        return success(message="Preference set")


@preference_api.route("/<int:im_id>/preference/<option>", methods=["DELETE"])
async def delete_preference(im_id, option):
    async with engine.session() as session:
        await session.execute(delete(UserPreferences).where(
            UserPreferences.user_id == im_id,
            UserPreferences.option == option
        ))
        await session.commit()
        return success(message="Preference deleted")
