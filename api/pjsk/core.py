import os
import pkgutil
import importlib
from quart import Blueprint
from types import ModuleType

from .db_engine import engine


def register_blueprints(bp: Blueprint):
    from . import routes

    routes_path = os.path.dirname(routes.__file__)
    for _, module_name, is_pkg in pkgutil.iter_modules([routes_path]):
        if not is_pkg:
            module_fullname = f"{routes.__name__}.{module_name}"
            module: ModuleType = importlib.import_module(module_fullname)
            for attr_name in dir(module):
                if attr_name.endswith("_api"):
                    obj = getattr(module, attr_name)
                    if isinstance(obj, Blueprint):
                        bp.register_blueprint(obj)


pjsk_api = Blueprint("pjsk_api", __name__, url_prefix="/pjsk")
register_blueprints(pjsk_api)


@pjsk_api.before_app_serving
async def init_db_engine():
    await engine.init_engine()


@pjsk_api.after_app_serving
async def shutdown_db_engine():
    await engine.shutdown_engine()
