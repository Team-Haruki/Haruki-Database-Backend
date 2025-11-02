package chunithm

import (
	"context"
	"fmt"
	"haruki-database/api"
	"haruki-database/config"
	harukiRedis "haruki-database/utils/redis"
	"net/http"

	entchuniMain "haruki-database/database/schema/chunithm/maindb"
	"haruki-database/database/schema/chunithm/maindb/chunithmmusicalias"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func getMusicIDByAlias(client *entchuniMain.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		aliasStr := c.Query("alias")
		if aliasStr == "" {
			return api.JSONResponse(c, http.StatusBadRequest, "alias is required")
		}

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "chunithm-music-alias")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		rows, err := client.ChunithmMusicAlias.
			Query().
			Where(chunithmmusicalias.AliasEQ(aliasStr)).
			All(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if len(rows) == 0 {
			return api.JSONResponse(c, http.StatusNotFound, "Alias not found")
		}

		ids := make([]int, len(rows))
		for i, r := range rows {
			ids[i] = r.MusicID
		}

		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", AliasToMusicIDResponse{
			Status:  200,
			Message: "success",
			Data:    ids,
		})
	}
}

func getAliasesByMusicID(client *entchuniMain.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		musicID, err := c.ParamsInt("music_id")
		if err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid music_id")
		}

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "chunithm-music-alias")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		rows, err := client.ChunithmMusicAlias.
			Query().
			Where(chunithmmusicalias.MusicIDEQ(musicID)).
			All(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		aliases := make([]string, len(rows))
		for i, r := range rows {
			aliases[i] = r.Alias
		}

		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", AllAliasesResponse{
			Data: aliases,
		})
	}
}

func addMusicAlias(client *entchuniMain.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		musicID, err := c.ParamsInt("music_id")
		if err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid music_id")
		}

		var body MusicAliasSchema
		if err := c.BodyParser(&body); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid request body")
		}

		existing, _ := client.ChunithmMusicAlias.
			Query().
			Where(chunithmmusicalias.MusicIDEQ(musicID), chunithmmusicalias.AliasEQ(body.Alias)).
			First(ctx)
		if existing != nil {
			return api.JSONResponse(c, http.StatusConflict, "Alias already exists")
		}

		newAlias, err := client.ChunithmMusicAlias.
			Create().
			SetMusicID(musicID).
			SetAlias(body.Alias).
			Save(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		query := fmt.Sprintf("alias=%s", newAlias.Alias)
		_ = harukiRedis.ClearCache(ctx, redisClient, "chunithm-music-alias", fmt.Sprintf("/chunithm/alias/%d", musicID), nil)
		_ = harukiRedis.ClearCache(ctx, redisClient, "chunithm-music-alias", "/chunithm/alias/music-id", &query)

		return api.JSONResponse(c, http.StatusOK, "Alias added", &MusicAliasSchema{ID: newAlias.ID, Alias: newAlias.Alias})
	}
}

func deleteMusicAlias(client *entchuniMain.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		musicID, err := c.ParamsInt("music_id")
		if err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid music_id")
		}

		var body MusicAliasSchema
		if err := c.BodyParser(&body); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid request body")
		}

		deleted, err := client.ChunithmMusicAlias.
			Delete().
			Where(chunithmmusicalias.MusicIDEQ(musicID), chunithmmusicalias.AliasEQ(body.Alias)).
			Exec(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if deleted == 0 {
			return api.JSONResponse(c, http.StatusNotFound, "Alias not found")
		}

		query := fmt.Sprintf("alias=%s", body.Alias)
		_ = harukiRedis.ClearCache(ctx, redisClient, "chunithm-music-alias", fmt.Sprintf("/chunithm/alias/%d", musicID), nil)
		_ = harukiRedis.ClearCache(ctx, redisClient, "chunithm-music-alias", "/chunithm/alias/music-id", &query)
		return api.JSONResponse(c, http.StatusOK, "Alias deleted")
	}
}

func registerAliasRoutes(router fiber.Router, client *entchuniMain.Client, redisClient *redis.Client) {
	r := router.Group("/alias")

	r.Get("/music-id", getMusicIDByAlias(client, redisClient))
	r.Get("/:music_id", getAliasesByMusicID(client, redisClient))
	r.Post("/:music_id/add", api.VerifyAPIAuthorization(), addMusicAlias(client, redisClient))
	r.Delete("/:music_id", api.VerifyAPIAuthorization(), deleteMusicAlias(client, redisClient))
}
