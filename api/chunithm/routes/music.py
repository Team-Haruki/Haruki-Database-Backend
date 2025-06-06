from quart import Blueprint, request
from sqlalchemy import select, desc

from ..db_engine import music_engine as engine
from ..schema import (
    MusicInfoSchema,
    MusicDifficultySchema,
    MusicBatchItemSchema,
    MusicBatchRequestSchema,
    ChartDataSchema,
)
from api.utils import success, error
from modules.sql.tables.chunithm import ChunithmMusicDifficulty, ChunithmMusic, ChunithmChartData

music_api = Blueprint("music_api", __name__, url_prefix="/music")


@music_api.route("/<int:music_id>/difficulty-info", methods=["GET"])
async def get_music_difficulty_info(music_id: int):
    version = request.args.get("version")
    if not version:
        return error("Missing version")

    async with engine.session() as session:
        stmt = select(ChunithmMusicDifficulty).where(
            ChunithmMusicDifficulty.music_id == music_id, ChunithmMusicDifficulty.version == version
        )
        result = await session.execute(stmt)
        record = result.scalar_one_or_none()

        if record:
            return success(MusicDifficultySchema.model_validate(record).model_dump())

        stmt_latest = (
            select(ChunithmMusicDifficulty)
            .where(ChunithmMusicDifficulty.music_id == music_id)
            .order_by(desc(ChunithmMusicDifficulty.version))
            .limit(1)
        )
        result = await session.execute(stmt_latest)
        latest = result.scalar_one_or_none()
        if latest:
            return success(MusicDifficultySchema.model_validate(latest).model_dump())

        return error("No difficulty data")


@music_api.route("/<int:music_id>/basic-info", methods=["GET"])
async def get_music_basic_info(music_id: int):
    async with engine.session() as session:
        stmt = select(ChunithmMusic).where(ChunithmMusic.music_id == music_id)
        result = await session.execute(stmt)
        music = result.scalar_one_or_none()

        if not music:
            return error("Music not found", code=404)

        return success(MusicInfoSchema.model_validate(music).model_dump())


@music_api.route("/<int:music_id>/chart-data", methods=["GET"])
async def get_music_chart_data(music_id: int):
    async with engine.session() as session:
        stmt = select(ChunithmChartData).where(ChunithmChartData.music_id == music_id)
        result = await session.execute(stmt)
        chart_rows = result.scalars().all()

        if not chart_rows:
            return error("No chart data found", code=404)

        chart_data = [ChartDataSchema.model_validate(row).model_dump() for row in chart_rows]
        return success(chart_data)


@music_api.route("/query-batch", methods=["POST"])
async def get_music_data_batch():
    data = await request.get_json()
    try:
        validated = MusicBatchRequestSchema.model_validate(data)
    except Exception as e:
        return error(f"Invalid request body: {str(e)}")

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

        return success(result)
