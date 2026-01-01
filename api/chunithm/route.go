package chunithm

import (
	"haruki-database/database/schema/chunithm/maindb"
	"haruki-database/database/schema/chunithm/music"
	"haruki-database/database/schema/users"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func RegisterChunithmRoutes(app fiber.Router, mainClient *maindb.Client, musicClient *music.Client, redisClient *redis.Client, usersClient *users.Client) {
	group := app.Group("/chunithm")
	registerAliasRoutes(group, mainClient, redisClient)
	registerBindingRoutes(group, mainClient, redisClient, usersClient)
	registerMusicRoutes(group, musicClient, redisClient)
}
