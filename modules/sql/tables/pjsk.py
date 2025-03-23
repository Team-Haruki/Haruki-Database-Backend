from sqlalchemy import Column, String, Integer, BigInteger

from .base import Base


class Binds(Base):
    __tablename__ = "binds"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    user_id = Column(String(50), nullable=False)
    game_id = Column(String(50), nullable=False)
    visible = Column(Integer, nullable=False)
    server = Column(String(10), nullable=False)


class UserPreferences(Base):
    __tablename__ = "user_preferences"
    user_id = Column(BigInteger, primary_key=True)
    option = Column(String(50), primary_key=True)
    value = Column(String(50), nullable=False)


class MusicAliases(Base):
    __tablename__ = "music_aliases"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    music_id = Column(Integer, nullable=False)
    alias = Column(String(100), nullable=False)


class GroupMusicAliases(Base):
    __tablename__ = "group_music_aliases"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    group_id = Column(String(50), nullable=False)
    music_id = Column(Integer, nullable=False)
    alias = Column(String(100), nullable=False)


class CharacterAliases(Base):
    __tablename__ = "character_aliases"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    character_id = Column(Integer, nullable=False)
    alias = Column(String(100), nullable=False)


class GroupCharacterAliases(Base):
    __tablename__ = "group_character_aliases"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    group_id = Column(String(50), nullable=False)
    character_id = Column(Integer, nullable=False)
    alias = Column(String(100), nullable=False)
