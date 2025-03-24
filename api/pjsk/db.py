from modules.sql.engine import DatabaseEngine
from configs.pjsk import PJSK_DB_HOST, PJSK_DB_PORT, PJSK_DB_USER, PJSK_DB_PASS, PJSK_DB_NAME

engine = DatabaseEngine(PJSK_DB_HOST, PJSK_DB_PORT, PJSK_DB_USER, PJSK_DB_PASS, PJSK_DB_NAME)
