from sqlalchemy import update, and_
from fastapi import APIRouter, Depends
from fastapi_cache import FastAPICache
from fastapi_cache.decorator import cache

from modules.exceptions import APIException
from modules.schemas.response import APIResponse
from modules.sql.tables.pjsk import UserPreference
from modules.schemas.pjsk import UserPreferenceSchema, UserPreferenceResultSchema
from utils import pjsk_engine as engine, parse_json_body, verify_api_auth

preference_api = APIRouter(prefix="/{platform}/user", tags=["user_preference"])


@preference_api.get(
    "/{im_id}/preference",
    response_model=UserPreferenceResultSchema,
    summary="获取全部偏好设置",
    description="获取指定平台下用户的所有偏好项",
)
@cache(expire=300)
async def get_preferences(platform: str, im_id: str, _: None = Depends(verify_api_auth)) -> UserPreferenceResultSchema:
    prefs = await engine.select(
        UserPreference, and_(UserPreference.platform == platform, UserPreference.im_id == im_id)
    )
    if prefs:
        return UserPreferenceResultSchema(options=[UserPreferenceSchema(option=p.option, value=p.value) for p in prefs])
    raise APIException(status=404, message="Preference not found")


@preference_api.get(
    "/{im_id}/preference/{option}",
    response_model=UserPreferenceResultSchema,
    summary="获取指定偏好项",
    description="获取指定平台下用户的某个偏好项",
)
@cache(expire=300)
async def get_preference_option(
    platform: str, im_id: str, option: str, _: None = Depends(verify_api_auth)
) -> UserPreferenceResultSchema:
    pref = await engine.select(
        UserPreference,
        and_(UserPreference.platform == platform, UserPreference.im_id == im_id, UserPreference.option == option),
        one_result=True,
    )
    if pref:
        return UserPreferenceResultSchema(option=UserPreferenceSchema.model_validate(pref))
    raise APIException(status=404, message="Preference not found")


@preference_api.put(
    "/{im_id}/preference",
    response_model=APIResponse,
    summary="设置或更新偏好项",
    description="设置用户的偏好项，如该项不存在则添加，存在则更新",
)
async def set_preference(
    platform: str,
    im_id: str,
    data: UserPreferenceSchema = Depends(parse_json_body(engine, UserPreferenceSchema)),
    _: None = Depends(verify_api_auth),
) -> APIResponse:
    async with engine.session() as session:
        stmt = (
            update(UserPreference)
            .where(
                UserPreference.platform == platform, UserPreference.im_id == im_id, UserPreference.option == data.option
            )
            .values(value=data.value)
        )
        result = await session.execute(stmt)
        if result.rowcount == 0:
            session.add(UserPreference(im_id=im_id, option=data.option, value=data.value))
        await session.commit()
        await FastAPICache.clear(namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/preference")
        await FastAPICache.clear(namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/preference/{data.option}")
    return APIResponse(message="Preference updated")


@preference_api.delete(
    "/{im_id}/preference/{option}",
    response_model=APIResponse,
    summary="删除偏好项",
    description="删除指定用户在某平台下的某个偏好项",
)
async def delete_preference(platform: str, im_id: str, option: str, _: None = Depends(verify_api_auth)) -> APIResponse:
    await engine.delete(
        UserPreference,
        and_(UserPreference.platform == platform, UserPreference.im_id == im_id, UserPreference.option == option),
    )
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/preference")
    await FastAPICache.clear(namespace="fastapi-cache", key=f"/{platform}/user/{im_id}/preference/{option}")
    return APIResponse(message="Preference deleted")
