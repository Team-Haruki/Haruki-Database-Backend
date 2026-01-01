package pjsk

import (
	"haruki-database/database/schema/pjsk"
	"haruki-database/database/schema/users"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func RegisterPJSKRoutes(app *fiber.App, client *pjsk.Client, redisClient *redis.Client, usersClient *users.Client) {
	group := app.Group("/pjsk")
	registerAliasRoutes(group, client, redisClient, usersClient)
	registerPreferenceRoutes(group, client, redisClient, usersClient)
	registerBindingRoutes(group, client, redisClient, usersClient)
}
