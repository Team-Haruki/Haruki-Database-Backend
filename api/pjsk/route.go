package pjsk

import (
	"haruki-database/database/schema/pjsk"

	"github.com/gofiber/fiber/v2"
)

func RegisterPJSKRoutes(app *fiber.App, client *pjsk.Client) {
	group := app.Group("/pjsk")
	RegisterAliasRoutes(group, client)
	RegisterPreferenceRoutes(group, client)
	RegisterBindingRoutes(group, client)
}
