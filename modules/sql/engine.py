from typing import Optional
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession

from .tables.base import Base


class DatabaseEngine:
    def __init__(self, url_scheme) -> None:
        self._engine = create_async_engine(url_scheme, echo=False, future=True)
        self._session_maker = async_sessionmaker(self._engine, expire_on_commit=False)

    async def init_engine(self):
        async with self._engine.begin() as conn:
            await conn.run_sync(Base.metadata.create_all)

    @asynccontextmanager
    async def session(self) -> AsyncGenerator[Optional[AsyncSession]]:
        async with self._session_maker() as _session:
            yield _session
