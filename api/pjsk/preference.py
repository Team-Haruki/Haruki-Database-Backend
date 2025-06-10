from sqlalchemy import update, and_
from fastapi import APIRouter, Depends
from fastapi_cache.decorator import cache


from modules.exceptions import APIException
from modules.schemas.response import APIResponse
from modules.sql.tables.pjsk import UserPreference
from utils import pjsk_engine as engine, parse_json_body, verify_api_auth
from modules.schemas.pjsk import UserPreferenceSchema, UserPreferenceResponse
from modules.cache_helpers import ORJsonCoder, cache_key_builder, clear_cache_by_path


preference_api = APIRouter(prefix="/{platform}/user", tags=["PJSK-User-Preference-API"])


@preference_api.get(
    "/{im_id}/preference",
    response_model=UserPreferenceResponse,
    summary="获取全部偏好设置",
    description="获取指定平台下用户的所有偏好项",
    dependencies=[Depends(verify_api_auth)],
)
@cache(namespace="pjsk_user_preference", coder=ORJsonCoder, expire=300, key_builder=cache_key_builder)  # type: ignore
async def get_preferences(platform: str, im_id: str) -> UserPreferenceResponse:
    prefs = await engine.select(
        UserPreference, and_(UserPreference.platform == platform, UserPreference.im_id == im_id)
    )
    if prefs:
        return UserPreferenceResponse(options=[UserPreferenceSchema(option=p.option, value=p.value) for p in prefs])
    raise APIException(status=404, message="Preference not found")


@preference_api.get(
    "/{im_id}/preference/{option}",
    response_model=UserPreferenceResponse,
    summary="获取指定偏好项",
    description="获取指定平台下用户的某个偏好项",
    dependencies=[Depends(verify_api_auth)],
)
@cache(namespace="pjsk_user_preference", coder=ORJsonCoder, expire=300, key_builder=cache_key_builder)  # type: ignore
async def get_preference_option(platform: str, im_id: str, option: str) -> UserPreferenceResponse:
    pref = await engine.select(
        UserPreference,
        and_(UserPreference.platform == platform, UserPreference.im_id == im_id, UserPreference.option == option),
        one_result=True,
    )
    if pref:
        return UserPreferenceResponse(option=UserPreferenceSchema.model_validate(pref))
    raise APIException(status=404, message="Preference not found")


@preference_api.put(
    "/{im_id}/preference",
    response_model=APIResponse,
    summary="设置或更新偏好项",
    description="设置用户的偏好项，如该项不存在则添加，存在则更新",
    dependencies=[Depends(verify_api_auth)],
)
async def set_preference(
    platform: str,
    im_id: str,
    data: UserPreferenceSchema = Depends(parse_json_body(engine, UserPreferenceSchema)),
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
            session.add(UserPreference(platform=platform, im_id=im_id, option=data.option, value=data.value))
        await session.commit()
        await clear_cache_by_path("pjsk_user_preference", f"/{platform}/user/{im_id}/preference")
        await clear_cache_by_path("pjsk_user_preference", f"/{platform}/user/{im_id}/preference/{data.option}")
    return APIResponse(message="Preference updated")


@preference_api.delete(
    "/{im_id}/preference/{option}",
    response_model=APIResponse,
    summary="删除偏好项",
    description="删除指定用户在某平台下的某个偏好项",
    dependencies=[Depends(verify_api_auth)],
)
async def delete_preference(platform: str, im_id: str, option: str) -> APIResponse:
    await engine.delete(
        UserPreference,
        and_(UserPreference.platform == platform, UserPreference.im_id == im_id, UserPreference.option == option),
    )
    await clear_cache_by_path("pjsk_user_preference", f"/{platform}/user/{im_id}/preference")
    await clear_cache_by_path("pjsk_user_preference", f"/{platform}/user/{im_id}/preference/{option}")
    return APIResponse(message="Preference deleted")
