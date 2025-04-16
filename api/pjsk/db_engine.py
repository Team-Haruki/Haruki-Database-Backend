from typing import Optional
from modules.sql.engine import DatabaseEngine
from configs.pjsk import PJSK_DB_URL

engine: Optional[DatabaseEngine] = DatabaseEngine(PJSK_DB_URL)
