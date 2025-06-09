from datetime import date
from pydantic import BaseModel, ConfigDict
from pydantic.generics import GenericModel
from typing import Optional, List, Union, TypeVar, Generic

T = TypeVar("T")


class ResponseWithData(GenericModel, Generic[T]):
    status: Optional[int] = 200
    message: Optional[str] = "success"
    data: Optional[T] = None


class MusicInfoSchema(BaseModel):
    music_id: int
    title: str
    artist: str
    category: str
    version: Optional[str] = None
    releaseDate: Optional[date] = None
    isDeleted: bool
    deletedVersion: Optional[str] = None

    model_config = ConfigDict(from_attributes=True)


class MusicDifficultySchema(BaseModel):
    music_id: int
    version: str
    diff0_const: Optional[float] = None
    diff1_const: Optional[float] = None
    diff2_const: Optional[float] = None
    diff3_const: Optional[float] = None
    diff4_const: Optional[float] = None

    model_config = ConfigDict(from_attributes=True)


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


class DefaultServerSchema(BaseModel):
    im_id: str
    platform: str
    server: str

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


class MusicAliasSchema(BaseModel):
    id: Optional[int] = None
    alias: str


class BindingSchema(BaseModel):
    im_id: str
    platform: str
    server: Optional[str] = None
    aime_id: Optional[str] = None

    model_config = ConfigDict(from_attributes=True)


class MusicDifficultyResponse(ResponseWithData[MusicDifficultySchema]):
    def __init__(self, **data):
        super().__init__(**data)


class MusicInfoResponse(ResponseWithData[MusicInfoSchema]):
    def __init__(self, **data):
        super().__init__(**data)


class AllMusicResponse(ResponseWithData[List[MusicInfoSchema]]):
    def __init__(self, **data):
        super().__init__(**data)


class ChartDataResponse(ResponseWithData[ChartDataSchema]):
    def __init__(self, **data):
        super().__init__(**data)


class DefaultServerResponse(ResponseWithData[DefaultServerSchema]):
    def __init__(self, **data):
        super().__init__(**data)


class BindingResponse(ResponseWithData[BindingSchema]):
    def __init__(self, **data):
        super().__init__(**data)


class AliasToMusicIDResponse(ResponseWithData[List[int]]):
    def __init__(self, **data):
        super().__init__(**data)


class AllAliasesResponse(ResponseWithData[List[str]]):
    def __init__(self, **data):
        super().__init__(**data)


class AddAliasResponse(ResponseWithData[MusicAliasSchema]):
    def __init__(self, **data):
        super().__init__(**data)
