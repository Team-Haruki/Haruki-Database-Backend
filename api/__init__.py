from .chunithm.core import chunithm_api
from .pjsk.core import pjsk_api

from configs.pjsk import PJSK_ENABLED
from configs.chunithm import CHUNITHM_ENABLED

apis = []
if PJSK_ENABLED:
    apis.append(pjsk_api)
if CHUNITHM_ENABLED:
    apis.append(chunithm_api)
