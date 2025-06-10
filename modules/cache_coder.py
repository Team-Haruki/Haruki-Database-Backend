import orjson
from typing import Any
from fastapi_cache import Coder
from fastapi.encoders import jsonable_encoder


class ORJsonCoder(Coder):
    @classmethod
    def encode(cls, value: Any) -> bytes:
        return orjson.dumps(
            value,
            default=jsonable_encoder,
            option=orjson.OPT_NON_STR_KEYS | orjson.OPT_SERIALIZE_NUMPY,
        )

    @classmethod
    def decode(cls, value: bytes) -> Any:
        return orjson.loads(value)
