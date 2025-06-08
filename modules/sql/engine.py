from sqlalchemy import select, delete
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from sqlalchemy.orm import InstrumentedAttribute
from typing import Optional, Callable, Union, List, Type, TypeVar
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession

from .tables.base import Base

T = TypeVar("T")


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

    async def shutdown_engine(self) -> None:
        await self._engine.dispose()

    async def select_data(
        self, target: Union[Type[T], InstrumentedAttribute], *conditions, one_result: bool = False
    ) -> Optional[Union[T, List[T]]]:
        async with self.session() as session:
            stmt = select(target)
            if conditions:
                stmt = stmt.where(*conditions)
            result = await session.execute(stmt)
            if one_result:
                return result.scalar_one_or_none()
            return result.scalars().all()

    async def delete_data(self, target: Union[Type[Base], InstrumentedAttribute], *conditions) -> Callable[[], int]:
        async with self.session() as session:
            stmt = delete(target)
            if conditions:
                stmt = stmt.where(*conditions)
            result = await session.execute(stmt)
            await session.commit()
            return result.rowcount

    async def add_data(self, instance: Base) -> None:
        async with self.session() as session:
            session.add(instance)
            await session.commit()
