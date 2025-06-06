from typing import Optional, List
from pydantic import BaseModel, ConfigDict


class MusicDifficultySchema(BaseModel):
    diff0_const: Optional[float] = None
    diff1_const: Optional[float] = None
    diff2_const: Optional[float] = None
    diff3_const: Optional[float] = None
    diff4_const: Optional[float] = None

    model_config = ConfigDict(from_attributes=True)


class MusicInfoSchema(BaseModel):
    title: str
    artist: str
    version: Optional[str] = None
    releaseDate: Optional[str] = None
    isDeleted: bool
    deletedVersion: Optional[str] = None

    model_config = ConfigDict(from_attributes=True)


class MusicBatchItemSchema(BaseModel):
    version: Optional[str] = None
    difficulty: List[Optional[float]]
    info: MusicInfoSchema


class MusicBatchResultSchema(BaseModel):
    __root__: dict[int, MusicBatchItemSchema]


class MusicBatchRequestSchema(BaseModel):
    music_ids: List[int]
    version: str


class ChartDataSchema(BaseModel):
    difficulty: int
    creator: Optional[str] = None
    bpm: Optional[float] = None
    tap_count: Optional[int] = None
    hold_count: Optional[int] = None
    slide_count: Optional[int] = None
    air_count: Optional[int] = None
    flick_count: Optional[int] = None
    total_count: Optional[int] = None

    model_config = ConfigDict(from_attributes=True)


class MusicTitleSchema(BaseModel):
    music_id: int
    title: str

    model_config = ConfigDict(from_attributes=True)


class MusicAliasSchema(BaseModel):
    alias: str


class BindingResultSchema(BaseModel):
    server: Optional[str] = None
    aime_id: Optional[str] = None
