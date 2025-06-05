import os
import pkgutil
import importlib
from quart import Blueprint
from types import ModuleType

from . import db_engine
from modules.sql.engine import DatabaseEngine
from configs.chunithm import CHUNITHM_BIND_DB_URL, CHUNITHM_MUSIC_DB_URL


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


chunithm_api = Blueprint("chunithm", __name__, url_prefix="/chunithm")
register_blueprints(chunithm_api)
db_engine.bind_engine = DatabaseEngine(CHUNITHM_BIND_DB_URL)
db_engine.music_engine = DatabaseEngine(CHUNITHM_MUSIC_DB_URL)
