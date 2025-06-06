from quart import jsonify
from typing import Optional
from modules.redis import RedisClient
from configs.redis import REDIS_HOST, REDIS_PORT, REDIS_PASSWORD

redis_client: Optional[RedisClient] = RedisClient(REDIS_HOST, REDIS_PORT, REDIS_PASSWORD)


def success(data=None, message="OK", code=200):
    return jsonify({"code": 0, "message": message, "data": data}), code


def error(message="Error", code=1):
    return jsonify({"code": code, "message": message}), 400
