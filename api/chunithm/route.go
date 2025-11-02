package chunithm

import (
	"haruki-database/database/schema/chunithm/maindb"
	"haruki-database/database/schema/chunithm/music"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func RegisterChunithmRoutes(app fiber.Router, mainClient *maindb.Client, musicClient *music.Client, redisClient *redis.Client) {
	group := app.Group("/chunithm")
	registerAliasRoutes(group, mainClient, redisClient)
	registerBindingRoutes(group, mainClient, redisClient)
	registerMusicRoutes(group, musicClient, redisClient)
}
