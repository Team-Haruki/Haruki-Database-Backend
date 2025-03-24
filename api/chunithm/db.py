from modules.sql.engine import DatabaseEngine
from configs.chunithm import (CHUNITHM_BIND_DB_HOST, CHUNITHM_BIND_DB_PORT, CHUNITHM_BIND_DB_USER,
                              CHUNITHM_BIND_DB_PASS, CHUNITHM_BIND_DB_NAME)
from configs.chunithm import (CHUNITHM_MUSIC_DB_HOST, CHUNITHM_MUSIC_DB_PORT, CHUNITHM_MUSIC_DB_USER,
                              CHUNITHM_MUSIC_DB_PASS, CHUNITHM_MUSIC_DB_NAME)

bind_engine = DatabaseEngine(CHUNITHM_BIND_DB_HOST, CHUNITHM_BIND_DB_PORT, CHUNITHM_BIND_DB_USER, CHUNITHM_BIND_DB_PASS,
                             CHUNITHM_BIND_DB_NAME)
music_engine = DatabaseEngine(CHUNITHM_MUSIC_DB_HOST, CHUNITHM_MUSIC_DB_PORT, CHUNITHM_MUSIC_DB_USER,
                              CHUNITHM_MUSIC_DB_PASS, CHUNITHM_MUSIC_DB_NAME)
