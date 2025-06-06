from typing import Optional, List
from pydantic import BaseModel, ConfigDict


class MusicDifficultySchema(BaseModel):
    diff0_const: Optional[float]
    diff1_const: Optional[float]
    diff2_const: Optional[float]
    diff3_const: Optional[float]
    diff4_const: Optional[float]

    model_config = ConfigDict(from_attributes=True)


class MusicInfoSchema(BaseModel):
    title: str
    artist: str
    version: Optional[str]
    releaseDate: Optional[str]
    isDeleted: bool
    deletedVersion: Optional[str]

    model_config = ConfigDict(from_attributes=True)


class MusicBatchItemSchema(BaseModel):
    version: Optional[str]
    difficulty: List[Optional[float]]
    info: MusicInfoSchema


class MusicBatchResultSchema(BaseModel):
    __root__: dict[int, MusicBatchItemSchema]


class MusicBatchRequestSchema(BaseModel):
    music_ids: List[int]
    version: str


class ChartDataSchema(BaseModel):
    difficulty: int
    creator: Optional[str]
    bpm: Optional[float]
    tap_count: Optional[int]
    hold_count: Optional[int]
    slide_count: Optional[int]
    air_count: Optional[int]
    flick_count: Optional[int]
    total_count: Optional[int]

    model_config = ConfigDict(from_attributes=True)
