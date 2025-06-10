from sqlalchemy import (
    Column,
    String,
    Integer,
    BigInteger,
    Float,
    Date,
    CheckConstraint,
    Numeric,
)

from .base import ChunithmMainBase, ChunithmMusicDBBase


class ChunithmBinding(ChunithmMainBase):
    __tablename__ = "bindings"
    im_id = Column(String(50), primary_key=True)
    platform = Column(String(50), primary_key=True)
    server = Column(String(10), primary_key=True)
    aime_id = Column(String(50), nullable=False)


class ChunithmDefaultServer(ChunithmMainBase):
    __tablename__ = "defaults"
    im_id = Column(String(50), primary_key=True)
    platform = Column(String(50), primary_key=True)
    server = Column(String(10), nullable=False)


class ChunithmMusicAlias(ChunithmMainBase):
    __tablename__ = "chunithm_aliases"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    music_id = Column(Integer, nullable=False)
    alias = Column(String(100), nullable=False)


class ChunithmChartData(ChunithmMusicDBBase):
    __tablename__ = "chart_data"
    music_id = Column(Integer, primary_key=True)
    difficulty = Column(Integer, primary_key=True)
    creator = Column(String(50), nullable=True)
    bpm = Column(Float, nullable=True)
    tap_count = Column(Integer, nullable=True)
    hold_count = Column(Integer, nullable=True)
    slide_count = Column(Integer, nullable=True)
    air_count = Column(Integer, nullable=True)
    flick_count = Column(Integer, nullable=True)
    total_count = Column(Integer, nullable=True)


class ChunithmMusic(ChunithmMusicDBBase):
    __tablename__ = "music"
    music_id = Column(Integer, primary_key=True)
    title = Column(String(255), nullable=False)
    artist = Column(String(255), nullable=False)
    category = Column(String(50), nullable=True)
    version = Column(String(10), nullable=True)
    release_date = Column(Date, nullable=True)
    is_deleted = Column(Integer, CheckConstraint("is_deleted IN (0, 1)"), default=0)
    deleted_version = Column(String(10), nullable=True)


class ChunithmMusicDifficulty(ChunithmMusicDBBase):
    __tablename__ = "music_difficulties"
    music_id = Column(Integer, primary_key=True)
    version = Column(String(10), primary_key=True)
    diff0_const = Column(Numeric(3, 1))
    diff1_const = Column(Numeric(3, 1))
    diff2_const = Column(Numeric(3, 1))
    diff3_const = Column(Numeric(3, 1))
    diff4_const = Column(Numeric(3, 1))
