package main

import (
	"context"
	"fmt"
	"io"
	"os"

	harukiConfig "haruki-database/config"
	harukiLogger "haruki-database/utils/logger"
	harukiRedis "haruki-database/utils/redis"

	"haruki-database/api/chunithm"
	"haruki-database/api/pjsk"

	chunithmMainDB "haruki-database/database/schema/chunithm/maindb"
	chunithmMusicDB "haruki-database/database/schema/chunithm/music"
	pjskDB "haruki-database/database/schema/pjsk"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var Version = "2.0.0-dev"

func main() {
	var logFile *os.File
	var loggerWriter io.Writer = os.Stdout
	if harukiConfig.Cfg.Backend.MainLogFile != "" {
		var err error
		logFile, err = os.OpenFile(harukiConfig.Cfg.Backend.MainLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			mainLogger := harukiLogger.NewLogger("Main", harukiConfig.Cfg.Backend.LogLevel, os.Stdout)
			mainLogger.Errorf("failed to open main log file: %v", err)
			os.Exit(1)
		}
		loggerWriter = io.MultiWriter(os.Stdout, logFile)
		defer logFile.Close()
	}
	mainLogger := harukiLogger.NewLogger("Main", harukiConfig.Cfg.Backend.LogLevel, loggerWriter)
	mainLogger.Infof(fmt.Sprintf("========================= Haruki Database Backend %s =========================", Version))
	mainLogger.Infof("Powered By Haruki Dev Team")
	mainLogger.Infof("Haruki Suite Backend Main Access Log Level: %s", harukiConfig.Cfg.Backend.LogLevel)
	mainLogger.Infof("Haruki Suite Backend Main Access Log Save Path: %s", harukiConfig.Cfg.Backend.MainLogFile)
	mainLogger.Infof("Go Fiber Access Log Save Path: %s", harukiConfig.Cfg.Backend.AccessLogPath)
	redisClient := harukiRedis.NewRedisClient(harukiConfig.Cfg.Redis)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		mainLogger.Errorf("Failed to connect Redis: %v", err)
		os.Exit(1)
	}

	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024,
	})

	var accessLogFile *os.File
	if harukiConfig.Cfg.Backend.AccessLog != "" {
		loggerConfig := logger.Config{Format: harukiConfig.Cfg.Backend.AccessLog}
		if harukiConfig.Cfg.Backend.AccessLogPath != "" {
			var err error
			accessLogFile, err = os.OpenFile(harukiConfig.Cfg.Backend.AccessLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				mainLogger.Errorf("Failed to open access log file: %v", err)
				os.Exit(1)
			}
			loggerConfig.Output = accessLogFile
		}
		app.Use(logger.New(loggerConfig))
	}
	if accessLogFile != nil {
		defer accessLogFile.Close()
	}

	var chunithmMainClient *chunithmMainDB.Client
	var chunithmMusicClient *chunithmMusicDB.Client
	var pjskClient *pjskDB.Client

	if harukiConfig.Cfg.Chunithm.Enabled {
		var err error
		chunithmMainClient, err = chunithmMainDB.Open(harukiConfig.Cfg.Chunithm.BindingDBType, harukiConfig.Cfg.Chunithm.BindingDBURL)
		if err != nil {
			mainLogger.Errorf("Failed to connect to Chunithm main DB: %v", err)
			os.Exit(1)
		}
		chunithmMusicClient, err = chunithmMusicDB.Open(harukiConfig.Cfg.Chunithm.MusicDBType, harukiConfig.Cfg.Chunithm.MusicDBURL)
		if err != nil {
			mainLogger.Errorf("Failed to connect to Chunithm music DB: %v", err)
			os.Exit(1)
		}
		defer chunithmMainClient.Close()
		defer chunithmMusicClient.Close()
		chunithm.RegisterChunithmRoutes(app, chunithmMainClient, chunithmMusicClient, redisClient)
	}

	if harukiConfig.Cfg.PJSK.Enabled {
		var err error
		pjskClient, err = pjskDB.Open(harukiConfig.Cfg.PJSK.DBType, harukiConfig.Cfg.PJSK.DBURL)
		if err != nil {
			mainLogger.Errorf("Failed to connect to PJSK DB: %v", err)
			os.Exit(1)
		}
		defer pjskClient.Close()
		pjsk.RegisterPJSKRoutes(app, pjskClient, redisClient)
	}

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
