package main

import (
	"context"
	"fmt"
	"io"
	"os"

	harukiConfig "haruki-database/config"
	harukiLogger "haruki-database/utils/logger"
	harukiRedis "haruki-database/utils/redis"

	botAPI "haruki-database/api/bot"
	censorAPI "haruki-database/api/censor"
	chunithmAPI "haruki-database/api/chunithm"
	PJSKAPI "haruki-database/api/pjsk"
	censorTool "haruki-database/utils/censor"

	botDB "haruki-database/database/schema/bot"
	censorDB "haruki-database/database/schema/censor"
	chunithmMainDB "haruki-database/database/schema/chunithm/maindb"
	chunithmMusicDB "haruki-database/database/schema/chunithm/music"
	pjskDB "haruki-database/database/schema/pjsk"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	_ "modernc.org/sqlite"
)

var Version = "2.0.0-dev"

func main() {
	loggerWriter := setupLogging()
	mainLogger := harukiLogger.NewLogger("Main", harukiConfig.Cfg.Backend.LogLevel, loggerWriter)

	logStartupInfo(mainLogger)

	redisClient := initRedis(mainLogger)
	app := createFiberApp(mainLogger)

	chunithmMainClient, chunithmMusicClient := initChunithmIfEnabled(mainLogger, app, redisClient)
	pjskClient := initPJSKIfEnabled(mainLogger, app, redisClient)
	censorDBClient, _ := initCensor(mainLogger, app)
	botDBClient := initBot(mainLogger, app, redisClient)

	defer closeClients(chunithmMainClient, chunithmMusicClient, pjskClient, censorDBClient, botDBClient)

	startServer(mainLogger, app)
}

func setupLogging() io.Writer {
	var logFile *os.File
	loggerWriter := io.Writer(os.Stdout)
	harukiConfig.LoadConfig("haruki-db-configs.yaml")

	if harukiConfig.Cfg.Backend.MainLogFile != "" {
		var err error
		logFile, err = os.OpenFile(harukiConfig.Cfg.Backend.MainLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			mainLogger := harukiLogger.NewLogger("Main", harukiConfig.Cfg.Backend.LogLevel, os.Stdout)
			mainLogger.Errorf("failed to open main log file: %v", err)
			os.Exit(1)
		}
		loggerWriter = io.MultiWriter(os.Stdout, logFile)
	}
	return loggerWriter
}

func logStartupInfo(mainLogger *harukiLogger.Logger) {
	mainLogger.Infof(fmt.Sprintf("========================= Haruki Database Backend %s =========================", Version))
	mainLogger.Infof("Powered By Haruki Dev Team")
	mainLogger.Infof("Haruki Suite Backend Main Access Log Level: %s", harukiConfig.Cfg.Backend.LogLevel)
	mainLogger.Infof("Haruki Suite Backend Main Access Log Save Path: %s", harukiConfig.Cfg.Backend.MainLogFile)
	mainLogger.Infof("Go Fiber Access Log Save Path: %s", harukiConfig.Cfg.Backend.AccessLogPath)
}

func initRedis(mainLogger *harukiLogger.Logger) *redis.Client {
	redisClient := harukiRedis.NewRedisClient(harukiConfig.Cfg.Redis)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		mainLogger.Errorf("Failed to connect Redis: %v", err)
		os.Exit(1)
	}
	return redisClient
}

func createFiberApp(mainLogger *harukiLogger.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024,
	})

	if harukiConfig.Cfg.Backend.AccessLog != "" {
		loggerConfig := logger.Config{Format: harukiConfig.Cfg.Backend.AccessLog}
		if harukiConfig.Cfg.Backend.AccessLogPath != "" {
			accessLogFile, err := os.OpenFile(harukiConfig.Cfg.Backend.AccessLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				mainLogger.Errorf("Failed to open access log file: %v", err)
				os.Exit(1)
			}
			loggerConfig.Output = accessLogFile
		}
		app.Use(logger.New(loggerConfig))
	}
	return app
}

func initChunithmIfEnabled(mainLogger *harukiLogger.Logger, app *fiber.App, redisClient *redis.Client) (*chunithmMainDB.Client, *chunithmMusicDB.Client) {
	if !harukiConfig.Cfg.Chunithm.Enabled {
		return nil, nil
	}

	chunithmMainClient, err := chunithmMainDB.Open(harukiConfig.Cfg.Chunithm.BindingDBType, harukiConfig.Cfg.Chunithm.BindingDBURL)
	if err != nil {
		mainLogger.Errorf("Failed to connect to Chunithm main DB: %v", err)
		os.Exit(1)
	}
	if err := chunithmMainClient.Schema.Create(context.Background()); err != nil {
		mainLogger.Errorf("Failed to create schema for Chunithm main DB: %v", err)
		os.Exit(1)
	}

	chunithmMusicClient, err := chunithmMusicDB.Open(harukiConfig.Cfg.Chunithm.MusicDBType, harukiConfig.Cfg.Chunithm.MusicDBURL)
	if err != nil {
		mainLogger.Errorf("Failed to connect to Chunithm music DB: %v", err)
		os.Exit(1)
	}
	if err := chunithmMusicClient.Schema.Create(context.Background()); err != nil {
		mainLogger.Errorf("Failed to create schema for Chunithm music DB: %v", err)
		os.Exit(1)
	}

	chunithmAPI.RegisterChunithmRoutes(app, chunithmMainClient, chunithmMusicClient, redisClient)
	return chunithmMainClient, chunithmMusicClient
}

func initPJSKIfEnabled(mainLogger *harukiLogger.Logger, app *fiber.App, redisClient *redis.Client) *pjskDB.Client {
	if !harukiConfig.Cfg.PJSK.Enabled {
		return nil
	}

	pjskClient, err := pjskDB.Open(harukiConfig.Cfg.PJSK.DBType, harukiConfig.Cfg.PJSK.DBURL)
	if err != nil {
		mainLogger.Errorf("Failed to connect to PJSK DB: %v", err)
		os.Exit(1)
	}
	if err := pjskClient.Schema.Create(context.Background()); err != nil {
		mainLogger.Errorf("Failed to create schema for PJSK DB: %v", err)
		os.Exit(1)
	}

	PJSKAPI.RegisterPJSKRoutes(app, pjskClient, redisClient)
	return pjskClient
}

func initCensor(mainLogger *harukiLogger.Logger, app *fiber.App) (*censorDB.Client, *censorTool.Service) {
	censorDBClient, err := censorDB.Open(harukiConfig.Cfg.Censor.CensorDBType, harukiConfig.Cfg.Censor.CensorDBURL)
	if err != nil {
		mainLogger.Errorf("Failed to initialize Censor entgo client: %v", err)
		os.Exit(1)
	}

	censorService := censorTool.NewService(harukiConfig.Cfg.Censor.BaiduAPIKey, harukiConfig.Cfg.Censor.BaiduSecret, censorDBClient)
	censorAPI.RegisterCensorRoutes(app, censorService)

	return censorDBClient, censorService
}

func initBot(mainLogger *harukiLogger.Logger, app *fiber.App, redisClient *redis.Client) *botDB.Client {
	botDBClient, err := botDB.Open(harukiConfig.Cfg.HarukiBotDB.DBType, harukiConfig.Cfg.HarukiBotDB.DBURL)
	if err != nil {
		mainLogger.Errorf("Failed to initialize Censor entgo client: %v", err)
		os.Exit(1)
	}

	botAPI.RegisterBotRoutes(app, botDBClient, redisClient)
	return botDBClient
}

func closeClients(chunithmMainClient *chunithmMainDB.Client, chunithmMusicClient *chunithmMusicDB.Client,
	pjskClient *pjskDB.Client, censorDBClient *censorDB.Client, botDBClient *botDB.Client) {
	if chunithmMainClient != nil {
		_ = chunithmMainClient.Close()
	}
	if chunithmMusicClient != nil {
		_ = chunithmMusicClient.Close()
	}
	if pjskClient != nil {
		_ = pjskClient.Close()
	}
	if censorDBClient != nil {
		_ = censorDBClient.Close()
	}
	if botDBClient != nil {
		_ = botDBClient.Close()
	}
}

func startServer(mainLogger *harukiLogger.Logger, app *fiber.App) {
	addr := fmt.Sprintf("%s:%d", harukiConfig.Cfg.Backend.Host, harukiConfig.Cfg.Backend.Port)
	if harukiConfig.Cfg.Backend.SSL {
		if err := app.ListenTLS(addr, harukiConfig.Cfg.Backend.SSLCert, harukiConfig.Cfg.Backend.SSLKey); err != nil {
			mainLogger.Errorf("Failed to start HTTPS server: %v", err)
			os.Exit(1)
		}
	} else {
		if err := app.Listen(addr); err != nil {
			mainLogger.Errorf("Failed to start HTTP server: %v", err)
			os.Exit(1)
		}
	}
}
