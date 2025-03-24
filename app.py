from quart import Quart, request, jsonify

from api import apis
from configs.app import ACCPET_AUTHORIZATION, ACCEPT_USER_AGENT

app = Quart(__name__)
for api in apis:
    app.register_blueprint(api)


@app.before_request
async def check_authorization():
    if ACCPET_AUTHORIZATION and request.headers.get('Authorization') != ACCPET_AUTHORIZATION:
        return jsonify({'error': 'Unauthorized'}), 401
    if ACCEPT_USER_AGENT and request.headers.get('User-Agent') != ACCEPT_USER_AGENT:
        return jsonify({'error': 'Invalid User Agent'}), 400
