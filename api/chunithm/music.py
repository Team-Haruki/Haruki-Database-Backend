from typing import Tuple, List
from sqlalchemy import select, desc, and_
from fastapi_cache.decorator import cache
from fastapi_limiter.depends import RateLimiter
from fastapi import APIRouter, Request, HTTPException, Depends, Query
from fastapi.responses import ORJSONResponse

from modules.exceptions import APIException
from utils import chunithm_music_engine as engine, parse_json_body
from modules.schemas.chunithm import (
    MusicInfoSchema,
    MusicDifficultySchema,
    MusicBatchItemSchema,
    MusicBatchRequestSchema,
    ChartDataSchema,
    AllMusicResponse,
    MusicBatchResultSchema, MusicDifficultyResponse, MusicInfoResponse, ChartDataResponse
)
from modules.sql.tables.chunithm import ChunithmMusicDifficulty, ChunithmMusic, ChunithmChartData

music_api = APIRouter(prefix="/music", tags=["music_api"])


@music_api.get("/all-music",
               response_model=AllMusicResponse,
               summary="查询所有乐曲",
               description="查询数据库中所有乐曲信息",
               dependencies=[Depends(RateLimiter(times=10, seconds=5))])
@cache(expire=300)
async def get_all_music() -> AllMusicResponse:
    async with engine.session() as session:
        stmt = select(ChunithmMusic)
        result = await session.execute(stmt)
        rows = result.scalars().all()
        return AllMusicResponse(data=[MusicInfoSchema.model_validate(row) for row in rows])


@music_api.get("/{music_id}/difficulty-info",
               response_model=MusicDifficultyResponse,
               summary="查询乐曲定数",
               description="根据提供的乐曲ID，返回对应乐曲基本信息",
               dependencies=[Depends(RateLimiter(times=10, seconds=1))])
@cache(expire=300)
async def get_music_difficulty_info(
    music_id: int, version: str = Query(..., description="查询版本，如2.30")
) -> MusicDifficultyResponse:
    record = await engine.select(
        ChunithmMusicDifficulty,
        and_(ChunithmMusicDifficulty.music_id == music_id, ChunithmMusicDifficulty.version == version),
        one_result=True,
    )
    if record:
        return MusicDifficultyResponse.model_validate(record)
    async with engine.session() as session:
        stmt_latest = (
            select(ChunithmMusicDifficulty)
            .where(ChunithmMusicDifficulty.music_id == music_id)
            .order_by(desc(ChunithmMusicDifficulty.version))
            .limit(1)
        )
        result = await session.execute(stmt_latest)
        latest = result.scalar_one_or_none()
        if latest:
            return MusicDifficultyResponse.model_validate(latest)
        raise APIException(status=404, message="No difficulty data")


@music_api.get("/{music_id}/basic-info",
               response_model=MusicInfoResponse,
               summary="查询乐曲基本信息",
               description="根据提供的乐曲ID，返回对应乐曲基本信息",
               dependencies=[Depends(RateLimiter(times=10, seconds=1))])
@cache(expire=300)
async def get_music_basic_info(music_id: int) -> MusicInfoResponse:
    music = await engine.select(ChunithmMusic, ChunithmMusic.music_id == music_id, one_result=True)
    if not music:
        raise APIException(status=404, message="Music not found")
    return MusicInfoResponse.model_validate(music)


@music_api.get("/{music_id}/chart-data",
               response_model=ChartDataResponse,
               summary="查询乐曲谱面数据",
               description="根据提供的乐曲ID，返回对应乐曲谱面数据",
               dependencies=[Depends(RateLimiter(times=10, seconds=1))])
@cache(expire=300)
async def get_music_chart_data(music_id: int) -> ChartDataResponse:
    chart_rows = await engine.select(ChunithmChartData, ChunithmChartData.music_id == music_id)
    if not chart_rows:
        raise APIException(status=404, message="No chart data found")
    chart_data = [ChartDataSchema.model_validate(row) for row in chart_rows]
    return ChartDataResponse(data=chart_data)


@music_api.post(
    "/query-batch",
    response_model=MusicBatchResultSchema,
    summary="批量查询乐曲信息与定数",
    description="根据提供的乐曲ID列表和版本号，返回对应的乐曲信息与定数",
    dependencies=[Depends(RateLimiter(times=3, seconds=1))],
)
async def get_music_data_batch(
    validated: MusicBatchRequestSchema = Depends(parse_json_body(engine, MusicBatchRequestSchema))
) -> MusicBatchResultSchema:
    music_ids = validated.music_ids
    version = validated.version
    async with engine.session() as session:
        music_stmt = select(ChunithmMusic).where(ChunithmMusic.music_id.in_(music_ids))
        difficulty_stmt = select(ChunithmMusicDifficulty).where(ChunithmMusicDifficulty.music_id.in_(music_ids))
        music_result = await session.execute(music_stmt)
        difficulty_result = await session.execute(difficulty_stmt)
        music_map = {m.music_id: m for m in music_result.scalars().all()}
        difficulty_map: dict[int, MusicDifficultySchema] = {}
        for d in sorted(difficulty_result.scalars().all(), key=lambda x: x.version, reverse=True):
            if d.music_id not in difficulty_map or d.version == version:
                difficulty_map[d.music_id] = MusicDifficultySchema.model_validate(d)
        result: dict[int, MusicBatchItemSchema] = {}
        for mid in music_ids:
            music = music_map.get(mid)
            difficulty = difficulty_map.get(mid)
            difficulty_list = [
                difficulty.diff0_const,
                difficulty.diff1_const,
                difficulty.diff2_const,
                difficulty.diff3_const,
                difficulty.diff4_const,
            ] if difficulty else [None] * 5
            if music:
                info = MusicInfoSchema.model_validate(music)
                version_val = music.version
            else:
                info = MusicInfoSchema(
                    music_id=mid,
                    title="Unknown",
                    artist="Unknown",
                    category="Unknown",
                    version=None,
                    releaseDate=None,
                    isDeleted=False,
                    deletedVersion=None,
                )
                version_val = None
            result[mid] = MusicBatchItemSchema(
                version=version_val,
                difficulty=difficulty_list,
                info=info,
            )
        return MusicBatchResultSchema(__root__=result)
