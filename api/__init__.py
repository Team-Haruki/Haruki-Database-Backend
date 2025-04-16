from configs.pjsk import PJSK_ENABLED
from configs.chunithm import CHUNITHM_ENABLED

apis = []
if PJSK_ENABLED:
    from .pjsk.core import pjsk_api
    apis.append(pjsk_api)
if CHUNITHM_ENABLED:
    from .chunithm.core import chunithm_api
    apis.append(chunithm_api)