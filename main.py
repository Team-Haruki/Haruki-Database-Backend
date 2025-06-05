import sys
import asyncio
import logging
import platform
import coloredlogs
from hypercorn.asyncio import serve
from hypercorn.config import Config

from app import app
from configs.app import LOG_FORMAT, FIELD_STYLE, HOST, PORT


async def run() -> None:
    config = Config()
    config.bind = [f"{HOST}:{PORT}"]
    config.loglevel = "info"
    config.access_log_format = "%(h)s %(r)s %(s)s %(b)s %(M)s"
    logging.basicConfig(level=logging.INFO)
    logger = logging.getLogger(__name__)
    coloredlogs.install(level="INFO", logger=logger, fmt=LOG_FORMAT, field_style=FIELD_STYLE)
    app.logger.setLevel(logging.INFO)
    await serve(app, config)


if __name__ == "__main__":
    if platform.system() == "Windows":
        asyncio.run(run())
    else:
        import uvloop

        python_version = sys.version_info
        if python_version >= (3, 11):
            uvloop.run(run())
        else:
            uvloop.install()
            asyncio.run(run())
