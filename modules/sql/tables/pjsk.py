from sqlalchemy import (
    Column,
    Integer,
    BigInteger,
    String,
    Boolean,
    DateTime,
    ForeignKey,
)
from sqlalchemy.orm import relationship

from .base import Base


class UserBinding(Base):
    __tablename__ = "user_bindings"
    id = Column(Integer, primary_key=True, autoincrement=True)
    im_id = Column(String, nullable=False, index=True)
    user_id = Column(String, nullable=False)
    server = Column(String, nullable=False)
    visible = Column(Boolean, default=True)
    default_refs = relationship(
        "UserDefaultBinding",
        back_populates="binding",
        cascade="all, delete",
        lazy="joined",
    )


class UserDefaultBinding(Base):
    __tablename__ = "user_default_bindings"
    im_id = Column(String, primary_key=True)
    server = Column(String, primary_key=True)  # 'jp', 'cn', ..., or 'default'
    bind_id = Column(Integer, ForeignKey("user_bindings.id", ondelete="CASCADE"), nullable=False)
    binding = relationship("UserBinding", back_populates="default_refs")


class UserPreferences(Base):
    __tablename__ = "user_preferences"
    user_id = Column(BigInteger, primary_key=True)
    option = Column(String(50), primary_key=True)
    value = Column(String(50), nullable=False)


class MusicAliases(Base):
    __tablename__ = "music_aliases"
    music_id = Column(Integer, nullable=False, primary_key=True)
    alias = Column(String(100), nullable=False, primary_key=True)


class GroupMusicAliases(Base):
    __tablename__ = "group_music_aliases"
    group_id = Column(String(50), nullable=False, primary_key=True)
    music_id = Column(Integer, nullable=False, primary_key=True)
    alias = Column(String(100), nullable=False, primary_key=True)


class CharacterAliases(Base):
    __tablename__ = "character_aliases"
    character_id = Column(Integer, nullable=False, primary_key=True)
    alias = Column(String(100), nullable=False, primary_key=True)


class GroupCharacterAliases(Base):
    __tablename__ = "group_character_aliases"
    group_id = Column(String(50), nullable=False, primary_key=True)
    character_id = Column(Integer, nullable=False, primary_key=True)
    alias = Column(String(100), nullable=False, primary_key=True)
