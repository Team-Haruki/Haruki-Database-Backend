import os
import pkgutil
import importlib
from quart import Blueprint
from types import ModuleType

from . import db_engine
from modules.sql.engine import DatabaseEngine
from configs.pjsk import PJSK_DB_HOST, PJSK_DB_PORT, PJSK_DB_USER, PJSK_DB_PASS, PJSK_DB_NAME


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


pjsk_api = Blueprint('pjsk_api', __name__, url_prefix='/pjsk')
register_blueprints(pjsk_api)
db_engine.engine = DatabaseEngine(PJSK_DB_HOST, PJSK_DB_PORT, PJSK_DB_USER, PJSK_DB_PASS, PJSK_DB_NAME)
