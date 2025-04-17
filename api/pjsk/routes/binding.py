from typing import Optional
from quart import Blueprint, request
from pydantic import ValidationError
from sqlalchemy import select, delete, update

from api.utils import error, success
from modules.sql.tables.pjsk import UserBinding, UserDefaultBinding
from ..db_engine import engine
from ..schema import AddBindingSchema, SetDefaultBindingSchema, UpdateBindingVisibilitySchema, BindingResult

user_binding_api = Blueprint("user_binding", __name__, url_prefix="/user")

@user_binding_api.route("/<im_id>/binding", methods=["GET"])
async def get_bindings(im_id):
    server: Optional[str] = request.args.get("server")
    async with engine.session() as session:
        stmt = select(UserBinding).where(UserBinding.im_id == im_id)
        if server:
            stmt = stmt.where(UserBinding.server == server)
        result = await session.execute(stmt)
        bindings = result.unique().scalars().all()
        return success([BindingResult.from_orm(b).model_dump() for b in bindings])


@user_binding_api.route("/<im_id>/binding", methods=["POST"])
async def add_binding(im_id):
    server = request.args.get("server", "jp")
    try:
        data = AddBindingSchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())
    async with engine.session() as session:
        exists_stmt = select(UserBinding).where(
            UserBinding.im_id == im_id,
            UserBinding.server == server,
            UserBinding.user_id == data.user_id
        )
        exists_result = await session.execute(exists_stmt)
        if exists_result.unique().scalar_one_or_none():
            return error("Binding already exists", code=409)
        binding = UserBinding(im_id=im_id, server=server, user_id=data.user_id, visible=data.visible)
        session.add(binding)
        await session.commit()
        return success({"id": binding.id})


@user_binding_api.route("/<im_id>/binding/default", methods=["GET"])
async def get_default_binding(im_id):
    server = request.args.get("server", "default")
    async with engine.session() as session:
        stmt = (
            select(UserBinding)
            .join(UserDefaultBinding)
            .where(
                UserDefaultBinding.im_id == im_id,
                UserDefaultBinding.server == server
            )
        )
        result = await session.execute(stmt)
        binding = result.unique().scalar_one_or_none()
        if not binding:
            return error(f"No default for server '{server}'" if server != "default" else "No global default set")
        return success(BindingResult.from_orm(binding).model_dump())


@user_binding_api.route("/<im_id>/bindings/default", methods=["PUT"])
async def set_default(im_id):
    server = request.args.get("server", "default")
    try:
        data = SetDefaultBindingSchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())
    async with engine.session() as session:
        bind_stmt = select(UserBinding).where(
            UserBinding.id == data.bind_id,
            UserBinding.im_id == im_id
        )
        result = await session.execute(bind_stmt)
        binding = result.scalar_one_or_none()
        if not binding:
            return error("Binding not found", code=403)
        await session.execute(
            delete(UserDefaultBinding).where(
                UserDefaultBinding.im_id == im_id,
                UserDefaultBinding.server == server
            )
        )
        session.add(UserDefaultBinding(im_id=im_id, server=server, bind_id=data.bind_id))
        await session.commit()
        return success(message=f"Set default for {server}")


@user_binding_api.route("/<im_id>/bindings/<int:bind_id>", methods=["PATCH"])
async def update_visibility(im_id, bind_id):
    try:
        data = UpdateBindingVisibilitySchema(**await request.get_json())
    except ValidationError as ve:
        return error(ve.errors())
    async with engine.session() as session:
        check_stmt = select(UserBinding).where(
            UserBinding.id == bind_id,
            UserBinding.im_id == im_id
        )
        result = await session.execute(check_stmt)
        binding = result.scalar_one_or_none()
        if not binding:
            return error("Binding not found", code=403)
        await session.execute(
            update(UserBinding)
            .where(UserBinding.id == bind_id)
            .values(visible=data.visible)
        )
        await session.commit()
        return success(message="Visibility updated")


@user_binding_api.route("/<im_id>/binding/<int:bind_id>", methods=["DELETE"])
async def delete_binding(im_id, bind_id):
    async with engine.session() as session:
        await session.execute(
            delete(UserDefaultBinding).where(UserDefaultBinding.im_id == im_id, UserDefaultBinding.bind_id == bind_id)
        )
        await session.execute(
            delete(UserBinding).where(UserBinding.id == bind_id, UserBinding.im_id == im_id)
        )
        await session.commit()
        return success(message="Binding deleted")