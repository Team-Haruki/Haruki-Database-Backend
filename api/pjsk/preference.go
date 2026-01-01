package pjsk

import (
	"context"
	"haruki-database/api"
	"haruki-database/config"
	"haruki-database/database/schema/pjsk"
	"haruki-database/database/schema/pjsk/userpreference"
	"haruki-database/database/schema/users"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

// ================= Preference Handlers =================

func (h *PreferenceHandler) GetAll(c fiber.Ctx) error {
	ctx := context.Background()
	harukiUserID := api.GetHarukiUserIDFromPath(c)
	if harukiUserID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid haruki_user_id")
	}
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSPreference)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	rows, err := h.svc.client.UserPreference.Query().
		Where(userpreference.HarukiUserIDEQ(harukiUserID)).
		All(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	if len(rows) == 0 {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrPreferenceNotFound)
	}
	out := make([]UserPreferenceSchema, len(rows))
	for i, r := range rows {
		out[i] = UserPreferenceSchema{Option: r.Option, Value: r.Value}
	}
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", UserPreferenceResponse{Options: out})
}
func (h *PreferenceHandler) Get(c fiber.Ctx) error {
	ctx := context.Background()
	harukiUserID := api.GetHarukiUserIDFromPath(c)
	if harukiUserID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid haruki_user_id")
	}
	option := c.Params("option")
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSPreference)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	row, err := h.svc.client.UserPreference.Query().
		Where(userpreference.HarukiUserIDEQ(harukiUserID), userpreference.OptionEQ(option)).
		First(ctx)
	if err != nil {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrPreferenceNotFound)
	}
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", UserPreferenceResponse{
		Option: &UserPreferenceSchema{Option: row.Option, Value: row.Value},
	})
}
func (h *PreferenceHandler) Update(c fiber.Ctx) error {
	ctx := context.Background()
	harukiUserID := api.GetHarukiUserIDFromPath(c)
	if harukiUserID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid haruki_user_id")
	}
	option := c.Params("option")
	var body UserPreferenceSchema
	if err := c.Bind().Body(&body); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	rows, err := h.svc.client.UserPreference.Update().
		Where(userpreference.HarukiUserIDEQ(harukiUserID), userpreference.OptionEQ(option)).
		SetValue(body.Value).
		Save(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	if rows == 0 {
		if _, err := h.svc.client.UserPreference.Create().
			SetHarukiUserID(harukiUserID).
			SetOption(option).
			SetValue(body.Value).
			Save(ctx); err != nil {
			return api.InternalError(c)
		}
	}
	h.svc.ClearCache(ctx, harukiUserID, option)
	return api.JSONResponse(c, fiber.StatusOK, "Preference updated")
}
func (h *PreferenceHandler) Delete(c fiber.Ctx) error {
	ctx := context.Background()
	harukiUserID := api.GetHarukiUserIDFromPath(c)
	if harukiUserID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid haruki_user_id")
	}
	option := c.Params("option")
	if _, err := h.svc.client.UserPreference.Delete().
		Where(userpreference.HarukiUserIDEQ(harukiUserID), userpreference.OptionEQ(option)).
		Exec(ctx); err != nil {
		return api.InternalError(c)
	}
	h.svc.ClearCache(ctx, harukiUserID, option)
	return api.JSONResponse(c, fiber.StatusOK, "Preference deleted")
}

// ================= Route Registration =================
func registerPreferenceRoutes(router fiber.Router, client *pjsk.Client, redisClient *redis.Client, usersClient *users.Client) {
	svc := NewPreferenceService(client, redisClient, usersClient)
	h := NewPreferenceHandler(svc)
	r := router.Group("/user/:haruki_user_id/preference", api.VerifyAPIAuthorization())
	r.Get("/", h.GetAll)
	r.Get("/:option", h.Get)
	r.Put("/:option", h.Update)
	r.Delete("/:option", h.Delete)
}
