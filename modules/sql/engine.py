from typing import Optional
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession


class DatabaseEngine:
    def __init__(self, host: str, port: int, user: str, password: str, db_name: str) -> None:
        self._engine = create_async_engine(
            f'mysql+aiomysql://{user}:{password}@{host}:{port}/{db_name}',
            echo=False,
            future=True
        )
        self._session_maker = async_sessionmaker(self._engine, expire_on_commit=False)

    @asynccontextmanager
    async def session(self) -> AsyncGenerator[Optional[AsyncSession]]:
        async with self._session_maker() as _session:
            yield _session
