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

from .base import Base


class ChunithmBind(Base):
    __tablename__ = "binds"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    user_id = Column(String(50), nullable=False)
    aime_id = Column(String(50), nullable=False)
    server = Column(String(10), nullable=False)


class ChunithmChartData(Base):
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


class ChunithmMusicAlias(Base):
    __tablename__ = "chunithm_aliases"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    music_id = Column(Integer, nullable=False)
    alias = Column(String(100), nullable=False)


class ChunithmMusic(Base):
    __tablename__ = "music"
    music_id = Column(Integer, primary_key=True)
    title = Column(String(255), nullable=False)
    artist = Column(String(255), nullable=False)
    category = Column(String(50), nullable=True)
    version = Column(String(10), nullable=True)
    release_date = Column(Date, nullable=True)
    is_deleted = Column(Integer, CheckConstraint("is_deleted IN (0, 1)"), default=0)
    deleted_version = Column(String(10), nullable=True)


class ChunithmMusicDifficulty(Base):
    __tablename__ = "music_difficulty"
    music_id = Column(Integer, primary_key=True)
    version = Column(String(10), primary_key=True)
    diff0_const = Column(Numeric(3, 1))
    diff1_const = Column(Numeric(3, 1))
    diff2_const = Column(Numeric(3, 1))
    diff3_const = Column(Numeric(3, 1))
    diff4_const = Column(Numeric(3, 1))
