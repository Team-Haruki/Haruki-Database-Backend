import orjson
from typing import Tuple
from pydantic import ValidationError
from sqlalchemy import select, update, delete
from quart import Blueprint, request, Response

from ..schema import UserPreferenceSchema
from ..db_engine import engine
from api.utils import redis_client
from api.utils import success, error
from modules.sql.tables.pjsk import UserPreference

preference_api = Blueprint("user_preference", __name__, url_prefix="/user")


@preference_api.get("/<im_id>/preference")
async def get_preferences(im_id: str) -> Tuple[Response, int]:
    cache_key = f"user_preferences:{im_id}"
    cache = await redis_client.get(cache_key)
    if cache:
        return success(orjson.loads(cache))
    async with engine.session() as session:
        result = await session.execute(select(UserPreference).where(UserPreference.im_id == im_id))
        prefs = result.scalars().all()
        result_data = [{"option": p.option, "value": p.value} for p in prefs]
        await redis_client.set(cache_key, orjson.dumps(result_data))
        return success(result_data)


@preference_api.get("/<im_id>/preference/<option>")
async def get_preference_option(im_id: str, option: str) -> Tuple[Response, int]:
    cache_key = f"user_preferences:{im_id}:{option}"
    cache = await redis_client.get(cache_key)
    if cache:
        return success(orjson.loads(cache))
    async with engine.session() as session:
        stmt = select(UserPreference).where(UserPreference.im_id == im_id, UserPreference.option == option)
        result = await session.execute(stmt)
        pref = result.scalar_one_or_none()
        if pref:
            result_data = UserPreferenceSchema.model_validate(pref).model_dump()
            await redis_client.set(cache_key, orjson.dumps(result_data))
            return success(result_data)
        return error("Preference not found.", code=404)


@preference_api.put("/<im_id>/preference")
async def set_preference(im_id: str) -> Tuple[Response, int]:
    try:
        data = UserPreferenceSchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())

    async with engine.session() as session:
        stmt = (
            update(UserPreference)
            .where(UserPreference.im_id == im_id, UserPreference.option == data.option)
            .values(value=data.value)
        )
        result = await session.execute(stmt)
        if result.rowcount == 0:
            session.add(UserPreference(im_id=im_id, option=data.option, value=data.value))
        await session.commit()
    await redis_client.delete(f"user_preferences:{im_id}")
    keys = await redis_client.keys(f"user_preferences:{im_id}:*")
    if keys:
        await redis_client.delete(*keys)
    return success(message="Preference set.")


@preference_api.route("/<int:im_id>/preference/<option>", methods=["DELETE"])
async def delete_preference(im_id: str, option: str) -> Tuple[Response, int]:
    async with engine.session() as session:
        await session.execute(
            delete(UserPreference).where(UserPreference.im_id == im_id, UserPreference.option == option)
        )
        await session.commit()
    await redis_client.delete(f"user_preferences:{im_id}")
    keys = await redis_client.keys(f"user_preferences:{im_id}:*")
    if keys:
        await redis_client.delete(*keys)
    return success(message="Preference deleted.")
