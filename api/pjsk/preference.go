package pjsk

import (
	"context"
	"fmt"
	"haruki-database/api"
	"haruki-database/config"
	harukiRedis "haruki-database/utils/redis"
	"net/http"

	"haruki-database/database/schema/pjsk"
	"haruki-database/database/schema/pjsk/userpreference"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func getUserPreferencesHandler(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "pjsk-user-preference")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		rows, err := client.UserPreference.
			Query().
			Where(
				userpreference.PlatformEQ(platform),
				userpreference.ImIDEQ(imID),
			).
			All(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if len(rows) == 0 {
			return api.JSONResponse(c, http.StatusNotFound, "Preference not found")
		}

		out := make([]UserPreferenceSchema, len(rows))
		for i, r := range rows {
			out[i] = UserPreferenceSchema{Option: r.Option, Value: r.Value}
		}
		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", UserPreferenceResponse{Options: out})
	}
}

func getUserPreferenceHandler(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")
		option := c.Params("option")

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "pjsk-user-preference")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		row, err := client.UserPreference.
			Query().
			Where(
				userpreference.PlatformEQ(platform),
				userpreference.ImIDEQ(imID),
				userpreference.OptionEQ(option),
			).
			First(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusNotFound, "Preference not found")
		}

		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", UserPreferenceResponse{
			Option: &UserPreferenceSchema{Option: row.Option, Value: row.Value},
		})
	}
}

func updateUserPreferenceHandler(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")
		option := c.Params("option")

		var body UserPreferenceSchema
		if err := c.Bind().Body(&body); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "Invalid request")
		}

		rows, err := client.UserPreference.
			Update().
			Where(
				userpreference.PlatformEQ(platform),
				userpreference.ImIDEQ(imID),
				userpreference.OptionEQ(option),
			).
			SetValue(body.Value).
			Save(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		if rows == 0 {
			_, err := client.UserPreference.
				Create().
				SetPlatform(platform).
				SetImID(imID).
				SetOption(option).
				SetValue(body.Value).
				Save(ctx)
			if err != nil {
				return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
			}
		}

		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-user-preference", fmt.Sprintf("/pjsk/%s/user/%s/preference", platform, imID), nil)
		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-user-preference", fmt.Sprintf("/pjsk/%s/user/%s/preference/%s", platform, imID, body.Option), nil)
		return api.JSONResponse(c, http.StatusOK, "Preference updated")
	}
}

func deleteUserPreferenceHandler(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")
		option := c.Params("option")

		_, err := client.UserPreference.
			Delete().
			Where(
				userpreference.PlatformEQ(platform),
				userpreference.ImIDEQ(imID),
				userpreference.OptionEQ(option),
			).
			Exec(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-user-preference", fmt.Sprintf("/pjsk/%s/user/%s/preference", platform, imID), nil)
		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-user-preference", fmt.Sprintf("/pjsk/%s/user/%s/preference/%s", platform, imID, option), nil)
		return api.JSONResponse(c, http.StatusOK, "Preference deleted")
	}
}

func registerPreferenceRoutes(router fiber.Router, client *pjsk.Client, redisClient *redis.Client) {
	r := router.Group("/:platform/user", api.VerifyAPIAuthorization())

	r.Get("/:im_id/preference", getUserPreferencesHandler(client, redisClient))
	r.Get("/:im_id/preference/:option", getUserPreferenceHandler(client, redisClient))
	r.Put("/:im_id/preference/:option", updateUserPreferenceHandler(client, redisClient))
	r.Delete("/:im_id/preference/:option", deleteUserPreferenceHandler(client, redisClient))
}
