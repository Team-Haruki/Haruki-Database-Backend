package pjsk

import (
	"haruki-database/database/schema/pjsk"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func RegisterPJSKRoutes(app *fiber.App, client *pjsk.Client, redisClient *redis.Client) {
	group := app.Group("/pjsk")
	RegisterAliasRoutes(group, client, redisClient)
	RegisterPreferenceRoutes(group, client, redisClient)
	RegisterBindingRoutes(group, client, redisClient)
}
