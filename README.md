# Haruki DB API

**Haruki DB API** is a companion project for [HarukiBot](https://github.com/Team-Haruki), providing API access to various SQL operations using `FastAPI`, `SQLAlchemy`, and `Pydantic`.  
It also utilizes `fastapi-cache2` for efficient caching to speed up API responses.

## Requirements
+ `MySQL` or `SQLite` (Not tested)
+ `Redis`

## How to Use

1. Edit all of the config files in directory `configs`.
2. Install [uv](https://github.com/astral-sh/uv) to manage and install project dependencies.
3. Run the following command to install dependencies:
   ```bash
   uv sync
   ```
4. (Optional) If you plan to use MySQL via aiomysql, install:
   ```bash
   uv add aiomysql
   ```
5. (Optional) If you plan to use SQLite via aiosqlite, install:
   ```bash
   uv add aiosqlite
   ```
6. (Optional) If you're on **Linux/macOS**, it's recommended to install [uvloop](https://github.com/MagicStack/uvloop) for better performance:
   ```bash
   uv add uvloop
   ```
7. If you need to change the listening address or other server settings, edit the `hypercorn.toml` file. If you have installed uvloop, uncomment the `worker_class` line in `hypercorn.toml` to enable it.
8. Finally, run the server using:
   ```bash
   hypercorn app:app --config hypercorn.toml
   ```

## License

This project is licensed under the MIT License.