CHUNITHM_ENABLED = True
CHUNITHM_MUSIC_DB_URL = "mysql+aiomysql://{user}:{password}@{host}:{port}/{db}".format(
    user="root", password="<PASSWORD>", host="127.0.0.1", port=3306, db="test"
)
CHUNITHM_BIND_DB_URL = "mysql+aiomysql://{user}:{password}@{host}:{port}/{db}".format(
    user="root", password="<PASSWORD>", host="127.0.0.1", port=3306, db="test"
)
