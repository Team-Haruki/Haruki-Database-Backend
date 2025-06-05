PJSK_ENABLED = True
PJSK_DB_URL = "mysql+aiomysql://{user}:{password}@{host}:{port}/{db}".format(
    user="root", password="<PASSWORD>", host="127.0.0.1", port=3306, db="test"
)
