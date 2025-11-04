package pjsk

import (
	"haruki-database/database/schema/pjsk"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func RegisterPJSKRoutes(app *fiber.App, client *pjsk.Client, redisClient *redis.Client) {
	group := app.Group("/pjsk")
	registerAliasRoutes(group, client, redisClient)
	registerPreferenceRoutes(group, client, redisClient)
	registerBindingRoutes(group, client, redisClient)
}
