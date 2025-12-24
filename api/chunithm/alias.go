package chunithm

import (
	"context"
	"haruki-database/api"
	"haruki-database/config"
	"haruki-database/database/schema/chunithm/maindb"
	"haruki-database/database/schema/chunithm/maindb/chunithmmusicalias"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func (h *AliasHandler) GetMusicIDByAlias(c fiber.Ctx) error {
	ctx := context.Background()
	aliasStr := c.Query("alias")
	if aliasStr == "" {
		return api.JSONResponse(c, fiber.StatusBadRequest, "alias is required")
	}
	if !api.ValidateAlias(aliasStr) {
		return api.JSONResponse(c, fiber.StatusBadRequest, "invalid alias")
	}
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSAlias)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	rows, err := h.svc.client.ChunithmMusicAlias.
		Query().
		Where(chunithmmusicalias.AliasEQ(aliasStr)).
		All(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	if len(rows) == 0 {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrAliasNotFound)
	}
	ids := extractMusicIDs(rows)
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", AliasToMusicIDResponse{MatchIDs: ids})
}

func (h *AliasHandler) GetAliasesByMusicID(c fiber.Ctx) error {
	ctx := context.Background()
	musicID := fiber.Params[int](c, "music_id", -1)
	if musicID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "invalid music_id")
	}
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSAlias)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	rows, err := h.svc.client.ChunithmMusicAlias.
		Query().
		Where(chunithmmusicalias.MusicIDEQ(musicID)).
		All(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	aliases := extractAliasStrings(rows)
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", AllAliasesResponse{Aliases: aliases})
}

func (h *AliasHandler) AddMusicAlias(c fiber.Ctx) error {
	ctx := context.Background()
	musicID := fiber.Params[int](c, "music_id", -1)
	if musicID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "invalid music_id")
	}
	var body AliasRequest
	if err := c.Bind().Body(&body); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	if !api.ValidateAlias(body.Alias) {
		return api.JSONResponse(c, fiber.StatusBadRequest, "invalid alias")
	}
	exists, _ := h.svc.client.ChunithmMusicAlias.
		Query().
		Where(chunithmmusicalias.MusicIDEQ(musicID), chunithmmusicalias.AliasEQ(body.Alias)).
		Exist(ctx)
	if exists {
		return api.JSONResponse(c, fiber.StatusConflict, api.ErrAlreadyExists)
	}
	newAlias, err := h.svc.client.ChunithmMusicAlias.
		Create().
		SetMusicID(musicID).
		SetAlias(body.Alias).
		Save(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	h.svc.ClearCache(ctx, musicID, body.Alias)
	return api.JSONResponse(c, fiber.StatusOK, "Alias added", MusicAliasSchema{ID: newAlias.ID, Alias: newAlias.Alias})
}

func (h *AliasHandler) DeleteMusicAlias(c fiber.Ctx) error {
	ctx := context.Background()
	musicID := fiber.Params[int](c, "music_id", -1)
	if musicID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "invalid music_id")
	}
	var body AliasRequest
	if err := c.Bind().Body(&body); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	deleted, err := h.svc.client.ChunithmMusicAlias.
		Delete().
		Where(chunithmmusicalias.MusicIDEQ(musicID), chunithmmusicalias.AliasEQ(body.Alias)).
		Exec(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	if deleted == 0 {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrAliasNotFound)
	}
	h.svc.ClearCache(ctx, musicID, body.Alias)
	return api.JSONResponse(c, fiber.StatusOK, "Alias deleted")
}

func registerAliasRoutes(router fiber.Router, client *maindb.Client, redisClient *redis.Client) {
	svc := NewAliasService(client, redisClient)
	h := NewAliasHandler(svc)
	r := router.Group("/alias")

	r.Get("/music-id", h.GetMusicIDByAlias)
	r.Get("/:music_id", h.GetAliasesByMusicID)
	r.Post("/:music_id", api.VerifyAPIAuthorization(), h.AddMusicAlias)
	r.Delete("/:music_id", api.VerifyAPIAuthorization(), h.DeleteMusicAlias)
}
