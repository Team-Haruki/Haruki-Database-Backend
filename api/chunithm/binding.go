package chunithm

import (
	"context"
	"fmt"
	"haruki-database/config"
	harukiRedis "haruki-database/utils/redis"
	"net/http"

	"haruki-database/api"
	entchuniMain "haruki-database/database/schema/chunithm/maindb"
	"haruki-database/database/schema/chunithm/maindb/chunithmbinding"
	"haruki-database/database/schema/chunithm/maindb/chunithmdefaultserver"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func RegisterBindingRoutes(router fiber.Router, client *entchuniMain.Client, redisClient *redis.Client) {
	r := router.Group("/:platform/user", api.VerifyAPIAuthorization())

	r.Get("/:im_id/default", func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")

		key, resp := api.CacheQuery(ctx, c, redisClient, "chunithm-binding")
		if resp != nil {
			return resp
		}

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

		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, *key, http.StatusOK, "ok", DefaultServerSchema{
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

		harukiRedis.ClearCache(ctx, redisClient, "chunithm-binding", fmt.Sprintf("/chunithm/%s/user/%s/default", platform, imID), nil)
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
		harukiRedis.ClearCache(ctx, redisClient, "chunithm-binding", fmt.Sprintf("/chunithm/%s/user/%s/default", platform, imID), nil)
		return api.JSONResponse(c, http.StatusOK, "Default server deleted")
	})

	r.Get("/:im_id/:server", func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")
		server := c.Params("server")

		key, resp := api.CacheQuery(ctx, c, redisClient, "chunithm-binding")
		if resp != nil {
			return resp
		}

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

		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, *key, http.StatusOK, "ok", BindingSchema{
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
		harukiRedis.ClearCache(ctx, redisClient, "chunithm-binding", fmt.Sprintf("/chunithm/%s/user/%s/%s", platform, imID, server), nil)
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
		harukiRedis.ClearCache(ctx, redisClient, "chunithm-binding", fmt.Sprintf("/chunithm/%s/user/%s/%s", platform, imID, server), nil)
		return api.JSONResponse(c, http.StatusOK, "Binding deleted")
	})
}
