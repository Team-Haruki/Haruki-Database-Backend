from quart import Quart

from api import apis

app = Quart(__name__)
for api in apis:
    app.register_blueprint(api)
