from sqlalchemy.orm import DeclarativeBase


class PjskBase(DeclarativeBase):
    __abstract__ = True


class ChunithmMainBase(DeclarativeBase):
    __abstract__ = True


class ChunithmMusicDBBase(DeclarativeBase):
    __abstract__ = True
