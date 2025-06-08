from typing import Tuple
from sqlalchemy import select, desc
from fastapi import APIRouter, Request, HTTPException
from fastapi.responses import ORJSONResponse
from fastapi import status


from utils import chunithm_music_engine as engine
from modules.schemas.chunithm import (
    MusicInfoSchema,
    MusicDifficultySchema,
    MusicBatchItemSchema,
    MusicBatchRequestSchema,
    ChartDataSchema,
    MusicTitleSchema,
)
from utils import success, error
from modules.sql.tables.chunithm import ChunithmMusicDifficulty, ChunithmMusic, ChunithmChartData

music_api = APIRouter(prefix="/music", tags=["music_api"])


@music_api.get("/all-music-titles")
async def get_all_music_titles():
    async with engine.session() as session:
        stmt = select(ChunithmMusic.music_id, ChunithmMusic.title)
        result = await session.execute(stmt)
        rows = result.all()
        return ORJSONResponse(content=success([MusicTitleSchema(music_id=r[0], title=r[1]).model_dump() for r in rows]))


@music_api.get("/{music_id}/difficulty-info")
async def get_music_difficulty_info(music_id: int, version: str = None):
    if not version:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Missing version")

    async with engine.session() as session:
        stmt = select(ChunithmMusicDifficulty).where(
            ChunithmMusicDifficulty.music_id == music_id, ChunithmMusicDifficulty.version == version
        )
        result = await session.execute(stmt)
        record = result.scalar_one_or_none()

        if record:
            return ORJSONResponse(content=success(MusicDifficultySchema.model_validate(record).model_dump()))

        stmt_latest = (
            select(ChunithmMusicDifficulty)
            .where(ChunithmMusicDifficulty.music_id == music_id)
            .order_by(desc(ChunithmMusicDifficulty.version))
            .limit(1)
        )
        result = await session.execute(stmt_latest)
        latest = result.scalar_one_or_none()
        if latest:
            return ORJSONResponse(content=success(MusicDifficultySchema.model_validate(latest).model_dump()))

        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="No difficulty data")


@music_api.get("/{music_id}/basic-info")
async def get_music_basic_info(music_id: int):
    async with engine.session() as session:
        stmt = select(ChunithmMusic).where(ChunithmMusic.music_id == music_id)
        result = await session.execute(stmt)
        music = result.scalar_one_or_none()

        if not music:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Music not found")

        return ORJSONResponse(content=success(MusicInfoSchema.model_validate(music).model_dump()))


@music_api.get("/{music_id}/chart-data")
async def get_music_chart_data(music_id: int):
    async with engine.session() as session:
        stmt = select(ChunithmChartData).where(ChunithmChartData.music_id == music_id)
        result = await session.execute(stmt)
        chart_rows = result.scalars().all()

        if not chart_rows:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="No chart data found")

        chart_data = [ChartDataSchema.model_validate(row).model_dump() for row in chart_rows]
        return ORJSONResponse(content=success(chart_data))


@music_api.post("/query-batch")
async def get_music_data_batch(request: Request):
    data = await request.json()
    try:
        validated = MusicBatchRequestSchema.model_validate(data)
    except Exception as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=f"Invalid request body: {str(e)}")

    music_ids = validated.music_ids
    version = validated.version

    async with engine.session() as session:
        music_stmt = select(ChunithmMusic).where(ChunithmMusic.music_id.in_(music_ids))
        difficulty_stmt = select(ChunithmMusicDifficulty).where(ChunithmMusicDifficulty.music_id.in_(music_ids))
        music_result = await session.execute(music_stmt)
        difficulty_result = await session.execute(difficulty_stmt)

        music_map = {m.music_id: m for m in music_result.scalars().all()}

        difficulty_rows = difficulty_result.scalars().all()
        diff_map = {}
        for d in sorted(difficulty_rows, key=lambda x: x.version, reverse=True):
            if d.music_id not in diff_map or d.version == version:
                diff_map[d.music_id] = MusicDifficultySchema.model_validate(d).model_dump().values()

        result = {}
        for mid in music_ids:
            music = music_map.get(mid)
            info = (
                MusicInfoSchema.model_validate(music).model_dump()
                if music
                else MusicInfoSchema(
                    title="Unknown",
                    artist="Unknown",
                    version=None,
                    releaseDate=None,
                    isDeleted=False,
                    deletedVersion=None,
                ).model_dump()
            )
            result[mid] = MusicBatchItemSchema(
                version=music.version if music else None,
                difficulty=list(diff_map.get(mid, [None] * 5)),
                info=MusicInfoSchema(**info),
            ).model_dump()

        return ORJSONResponse(content=success(result))
