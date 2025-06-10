from sqlalchemy import select, delete
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from sqlalchemy.orm import InstrumentedAttribute, DeclarativeBase
from typing import Optional, Callable, Union, List, Type, TypeVar
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession

T = TypeVar("T")
JoinTarget = TypeVar("JoinTarget")


class DatabaseEngine:
    def __init__(self, url_scheme, table_base: Type[DeclarativeBase]) -> None:
        self._engine = create_async_engine(url_scheme, echo=False, future=True)
        self._session_maker = async_sessionmaker(self._engine, expire_on_commit=False)
        self._table_base = table_base

    async def init_engine(self):
        async with self._engine.begin() as conn:
            await conn.run_sync(self._table_base.metadata.create_all)

    @asynccontextmanager
    async def session(self) -> AsyncGenerator[Optional[AsyncSession]]:
        async with self._session_maker() as _session:
            yield _session

    async def shutdown_engine(self) -> None:
        await self._engine.dispose()

    async def select(
        self, target: Union[Type[T], InstrumentedAttribute], *conditions, one_result: bool = False, unique: bool = False
    ) -> Optional[Union[T, List[T]]]:
        async with self.session() as session:
            stmt = select(target)
            if conditions:
                stmt = stmt.where(*conditions)
            result = await session.execute(stmt)
            if unique:
                result = result.unique()
            if one_result:
                return result.scalar_one_or_none()
            return result.scalars().all()

    async def select_with_join(
        self,
        target: Union[Type[T], InstrumentedAttribute],
        join_model: Union[Type[JoinTarget], InstrumentedAttribute],
        on_clause,
        *conditions,
        one_result: bool = False,
        unique: bool = False,
    ) -> Optional[Union[T, List[T]]]:
        async with self.session() as session:
            stmt = select(target).join(join_model, on_clause)
            if conditions:
                stmt = stmt.where(*conditions)
            result = await session.execute(stmt)
            if unique:
                result = result.unique()
            if one_result:
                return result.scalar_one_or_none()
            return result.scalars().all()

    async def delete(self, target: Union[Type[T], InstrumentedAttribute], *conditions) -> Callable[[], int]:
        async with self.session() as session:
            stmt = delete(target)
            if conditions:
                stmt = stmt.where(*conditions)
            result = await session.execute(stmt)
            await session.commit()
            return result.rowcount

    async def add(self, instance: T) -> Optional[T]:
        async with self.session() as session:
            session.add(instance)
            await session.commit()
            await session.refresh(instance)
            return instance
