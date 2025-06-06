from typing import Tuple
from sqlalchemy import select
from quart import Blueprint, Response
from ..db_engine import bind_engine as engine
from ..schema import BindingResultSchema
from api.utils import success, error
from modules.sql.tables.chunithm import ChunithmBind, ChunithmDefaultServer

binding_api = Blueprint("binding_api", __name__, url_prefix="/user")


@binding_api.get("/<im_id>/default")
async def get_default_server(im_id: str) -> Tuple[Response, int]:
    async with engine.session() as session:
        stmt = select(ChunithmDefaultServer.server).where(ChunithmDefaultServer.im_id == im_id)
        result = await session.execute(stmt)
        server = result.scalar_one_or_none()
        if server is None:
            return error("Default server not set", code=404)
        return success(BindingResultSchema(server=server).model_dump(exclude_none=True))


@binding_api.get("/<im_id>/<server>")
async def get_binding(im_id: str, server: str) -> Tuple[Response, int]:
    async with engine.session() as session:
        stmt = select(ChunithmBind.aime_id).where(
            ChunithmBind.im_id == im_id,
            ChunithmBind.server == server
        )
        result = await session.execute(stmt)
        aime_id = result.scalar_one_or_none()
        if aime_id is None:
            return error("Binding not found", code=404)
        return success(BindingResultSchema(aime_id=aime_id).model_dump(exclude_none=True))


@binding_api.put("/<im_id>/<server>/<aime_id>")
async def update_binding(im_id: str, server: str, aime_id: str) -> Tuple[Response, int]:
    async with engine.session() as session:
        stmt = select(ChunithmBind).where(
            ChunithmBind.im_id == im_id,
            ChunithmBind.server == server
        )
        result = await session.execute(stmt)
        existing = result.scalar_one_or_none()
        if existing:
            existing.aime_id = aime_id
        else:
            session.add(ChunithmBind(im_id=im_id, server=server, aime_id=aime_id))
        await session.commit()
        return success(message="Binding updated")


@binding_api.delete("/<im_id>/<server>/<aime_id>")
async def delete_binding(im_id: str, server: str, aime_id: str) -> Tuple[Response, int]:
    async with engine.session() as session:
        stmt = select(ChunithmBind).where(
            ChunithmBind.im_id == im_id,
            ChunithmBind.server == server,
            ChunithmBind.aime_id == aime_id
        )
        result = await session.execute(stmt)
        bind = result.scalar_one_or_none()
        if bind is None:
            return error("Binding not found", code=404)
        await session.delete(bind)
        await session.commit()
        return success("Binding deleted")
