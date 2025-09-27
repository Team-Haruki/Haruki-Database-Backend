package chunithm

import (
	"haruki-database/database/schema/chunithm/maindb"
	"haruki-database/database/schema/chunithm/music"

	"github.com/gofiber/fiber/v2"
)

func RegisterChunithmRoutes(app fiber.Router, mainClient *maindb.Client, musicClient *music.Client) {
	group := app.Group("/chunithm")
	RegisterAliasRoutes(group, mainClient)
	RegisterBindingRoutes(group, mainClient)
	RegisterMusicRoutes(group, musicClient)
}
