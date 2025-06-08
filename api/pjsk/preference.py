import orjson
from typing import Tuple
from pydantic import ValidationError
from sqlalchemy import select, update, delete
from fastapi import APIRouter, Request, HTTPException
from fastapi.responses import ORJSONResponse
from fastapi import status

from modules.schemas.pjsk import UserPreferenceSchema
from utils import pjsk_engine as engine
from utils import redis_client
from utils import success, error
from modules.sql.tables.pjsk import UserPreference

preference_api = APIRouter(prefix="/user", tags=["user_preference"])


@preference_api.get("/{im_id}/preference")
async def get_preferences(im_id: str):
    cache_key = f"user_preferences:{im_id}"
    cache = await redis_client.get(cache_key)
    if cache:
        return ORJSONResponse(content=success(orjson.loads(cache)))
    async with engine.session() as session:
        result = await session.execute(select(UserPreference).where(UserPreference.im_id == im_id))
        prefs = result.scalars().all()
        result_data = [{"option": p.option, "value": p.value} for p in prefs]
        await redis_client.set(cache_key, orjson.dumps(result_data))
        return ORJSONResponse(content=success(result_data))


@preference_api.get("/{im_id}/preference/{option}")
async def get_preference_option(im_id: str, option: str):
    cache_key = f"user_preferences:{im_id}:{option}"
    cache = await redis_client.get(cache_key)
    if cache:
        return ORJSONResponse(content=success(orjson.loads(cache)))
    async with engine.session() as session:
        stmt = select(UserPreference).where(UserPreference.im_id == im_id, UserPreference.option == option)
        result = await session.execute(stmt)
        pref = result.scalar_one_or_none()
        if pref:
            result_data = UserPreferenceSchema.model_validate(pref).model_dump()
            await redis_client.set(cache_key, orjson.dumps(result_data))
            return ORJSONResponse(content=success(result_data))
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Preference not found.")


@preference_api.put("/{im_id}/preference")
async def set_preference(im_id: str, request: Request):
    try:
        data = UserPreferenceSchema(**(await request.json()))
    except ValidationError as ve:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_ENTITY, detail=ve.errors())

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
    return ORJSONResponse(content=success(message="Preference set."))


@preference_api.delete("/{im_id}/preference/{option}")
async def delete_preference(im_id: str, option: str):
    async with engine.session() as session:
        await session.execute(
            delete(UserPreference).where(UserPreference.im_id == im_id, UserPreference.option == option)
        )
        await session.commit()
    await redis_client.delete(f"user_preferences:{im_id}")
    keys = await redis_client.keys(f"user_preferences:{im_id}:*")
    if keys:
        await redis_client.delete(*keys)
    return ORJSONResponse(content=success(message="Preference deleted."))
