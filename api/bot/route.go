package bot

import (
	ent "haruki-database/database/schema/bot"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func RegisterBotRoutes(app *fiber.App, dbClient *ent.Client, redisClient *redis.Client) {
	registerUserRoutes(app, dbClient, redisClient)
	registerStatisticsRoutes(app, dbClient)
}
