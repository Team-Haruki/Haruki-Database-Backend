package chunithm

import (
	"context"
	"net/http"

	"haruki-database/api"
	entchuniMain "haruki-database/database/schema/chunithm/maindb"
	"haruki-database/database/schema/chunithm/maindb/chunithmbinding"
	"haruki-database/database/schema/chunithm/maindb/chunithmdefaultserver"

	"github.com/gofiber/fiber/v2"
)

func RegisterBindingRoutes(router fiber.Router, client *entchuniMain.Client) {
	r := router.Group("/:platform/user", api.VerifyAPIAuthorization())

	r.Get("/:im_id/default", func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")

		row, err := client.ChunithmDefaultServer.
			Query().
			Where(
				chunithmdefaultserver.PlatformEQ(platform),
				chunithmdefaultserver.ImIDEQ(imID),
			).
			First(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusNotFound, "Default server not set")
		}

		return api.JSONResponse(c, http.StatusOK, "success", DefaultServerSchema{
			ImID:     row.ImID,
			Platform: row.Platform,
			Server:   row.Server,
		})
	})

	r.Put("/:im_id/default/:server", func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")
		server := c.Params("server")

		row, _ := client.ChunithmDefaultServer.
			Query().
			Where(chunithmdefaultserver.PlatformEQ(platform), chunithmdefaultserver.ImIDEQ(imID)).
			First(ctx)

		if row != nil {
			_, err := row.Update().SetServer(server).Save(ctx)
			if err != nil {
				return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
			}
		} else {
			_, err := client.ChunithmDefaultServer.
				Create().
				SetPlatform(platform).
				SetImID(imID).
				SetServer(server).
				Save(ctx)
			if err != nil {
				return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
			}
		}
		return api.JSONResponse(c, http.StatusOK, "Default server set")
	})

	r.Delete("/:im_id/default", func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")

		count, err := client.ChunithmDefaultServer.
			Delete().
			Where(chunithmdefaultserver.PlatformEQ(platform), chunithmdefaultserver.ImIDEQ(imID)).
			Exec(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if count == 0 {
			return api.JSONResponse(c, http.StatusNotFound, "Default server not set")
		}
		return api.JSONResponse(c, http.StatusOK, "Default server deleted")
	})

	r.Get("/:im_id/:server", func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")
		server := c.Params("server")

		row, err := client.ChunithmBinding.
			Query().
			Where(
				chunithmbinding.PlatformEQ(platform),
				chunithmbinding.ImIDEQ(imID),
				chunithmbinding.ServerEQ(server),
			).
			First(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusNotFound, "Binding not found")
		}

		return api.JSONResponse(c, http.StatusOK, "success", BindingSchema{
			ImID:     row.ImID,
			Platform: row.Platform,
			Server:   &row.Server,
			AimeID:   &row.AimeID,
		})
	})

	r.Put("/:im_id/:server/:aime_id", func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")
		server := c.Params("server")
		aimeID := c.Params("aime_id")

		row, _ := client.ChunithmBinding.
			Query().
			Where(
				chunithmbinding.ImIDEQ(imID),
				chunithmbinding.ServerEQ(server),
			).
			First(ctx)

		if row != nil {
			_, err := row.Update().SetAimeID(aimeID).Save(ctx)
			if err != nil {
				return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
			}
		} else {
			_, err := client.ChunithmBinding.
				Create().
				SetPlatform(platform).
				SetImID(imID).
				SetServer(server).
				SetAimeID(aimeID).
				Save(ctx)
			if err != nil {
				return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
			}
		}

		return api.JSONResponse(c, http.StatusOK, "Binding updated")
	})

	r.Delete("/:im_id/:server/:aime_id", func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")
		server := c.Params("server")
		aimeID := c.Params("aime_id")

		count, err := client.ChunithmBinding.
			Delete().
			Where(
				chunithmbinding.PlatformEQ(platform),
				chunithmbinding.ImIDEQ(imID),
				chunithmbinding.ServerEQ(server),
				chunithmbinding.AimeIDEQ(aimeID),
			).
			Exec(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if count == 0 {
			return api.JSONResponse(c, http.StatusNotFound, "Binding not found")
		}
		return api.JSONResponse(c, http.StatusOK, "Binding deleted")
	})
}
