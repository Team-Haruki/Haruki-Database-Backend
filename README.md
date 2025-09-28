# Haruki Database Backend

**Haruki Database Backend** is a companion project for [HarukiBot](https://github.com/Team-Haruki), providing API access to various SQL operations using `Fiber`, `EntGo`
It also utilizes `Redis` for efficient caching to speed up some APIs' response.

## Requirements
+ `MySQL` or `SQLite` (Not tested)
+ `Redis`

## How to Use

1. Go to release page to download `HarukiDatabaseBackend`
2. Download `haruki-db-configs.example.yaml`, and rename it to `haruki-db-configs.yaml`
3. Make a new directory or use an exists directory
4. Put `HarukiDatabaseBackend` and `haruki-db-configs.yaml` in the same directory
5. Edit `haruki-db-configs.yaml` and configure it
6. Open Terminal, and `cd` to the directory
7. Run `HarukiDatabaseBackend`

## License

This project is licensed under the MIT License.