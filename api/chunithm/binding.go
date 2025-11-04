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

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func getDefaultServer(client *entchuniMain.Client, redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "chunithm-binding")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
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

		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", DefaultServerSchema{
			ImID:     row.ImID,
			Platform: row.Platform,
			Server:   row.Server,
		})
	}
}

func setDefaultServer(client *entchuniMain.Client, redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
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

		_ = harukiRedis.ClearCache(ctx, redisClient, "chunithm-binding", fmt.Sprintf("/chunithm/%s/user/%s/default", platform, imID), nil)
		return api.JSONResponse(c, http.StatusOK, "Default server set")
	}
}

func deleteDefaultServer(client *entchuniMain.Client, redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
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
		_ = harukiRedis.ClearCache(ctx, redisClient, "chunithm-binding", fmt.Sprintf("/chunithm/%s/user/%s/default", platform, imID), nil)
		return api.JSONResponse(c, http.StatusOK, "Default server deleted")
	}
}

func getBinding(client *entchuniMain.Client, redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")
		server := c.Params("server")

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "chunithm-binding")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
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

		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", BindingSchema{
			ImID:     row.ImID,
			Platform: row.Platform,
			Server:   &row.Server,
			AimeID:   &row.AimeID,
		})
	}
}

func setBinding(client *entchuniMain.Client, redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
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
		_ = harukiRedis.ClearCache(ctx, redisClient, "chunithm-binding", fmt.Sprintf("/chunithm/%s/user/%s/%s", platform, imID, server), nil)
		return api.JSONResponse(c, http.StatusOK, "Binding updated")
	}
}

func deleteBinding(client *entchuniMain.Client, redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
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
		_ = harukiRedis.ClearCache(ctx, redisClient, "chunithm-binding", fmt.Sprintf("/chunithm/%s/user/%s/%s", platform, imID, server), nil)
		return api.JSONResponse(c, http.StatusOK, "Binding deleted")
	}
}

func registerBindingRoutes(router fiber.Router, client *entchuniMain.Client, redisClient *redis.Client) {
	r := router.Group("/:platform/user", api.VerifyAPIAuthorization())

	r.Get("/:im_id/default", getDefaultServer(client, redisClient))
	r.Put("/:im_id/default/:server", setDefaultServer(client, redisClient))
	r.Delete("/:im_id/default", deleteDefaultServer(client, redisClient))
	r.Get("/:im_id/:server", getBinding(client, redisClient))
	r.Put("/:im_id/:server/:aime_id", setBinding(client, redisClient))
	r.Delete("/:im_id/:server/:aime_id", deleteBinding(client, redisClient))
}
