from quart import jsonify


def success(data=None, message="OK"):
    return jsonify({"code": 0, "message": message, "data": data})


def error(message="Error", code=1):
    return jsonify({"code": code, "message": message}), 400
