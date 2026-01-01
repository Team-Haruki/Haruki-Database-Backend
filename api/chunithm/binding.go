package chunithm

import (
	"context"
	"haruki-database/api"
	"haruki-database/config"
	entchuniMain "haruki-database/database/schema/chunithm/maindb"
	"haruki-database/database/schema/chunithm/maindb/chunithmbinding"
	"haruki-database/database/schema/chunithm/maindb/chunithmdefaultserver"
	"haruki-database/database/schema/users"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

// ================= Binding Handlers =================

func (h *BindingHandler) GetDefaultServer(c fiber.Ctx) error {
	ctx := context.Background()
	userID := api.GetHarukiUserIDFromPath(c)
	if userID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid haruki_user_id")
	}

	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSBinding)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}

	row, err := h.svc.client.ChunithmDefaultServer.
		Query().
		Where(chunithmdefaultserver.HarukiUserIDEQ(userID)).
		First(ctx)
	if err != nil {
		return api.JSONResponse(c, fiber.StatusNotFound, "Default server not set")
	}

	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", DefaultServerSchema{
		UserID: row.HarukiUserID,
		Server: row.Server,
	})
}

func (h *BindingHandler) SetDefaultServer(c fiber.Ctx) error {
	ctx := context.Background()
	userID := api.GetHarukiUserIDFromPath(c)
	if userID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid haruki_user_id")
	}
	server := c.Params("server")

	row, _ := h.svc.client.ChunithmDefaultServer.
		Query().
		Where(chunithmdefaultserver.HarukiUserIDEQ(userID)).
		First(ctx)

	if row != nil {
		if _, err := row.Update().SetServer(server).Save(ctx); err != nil {
			return api.InternalError(c)
		}
	} else {
		if _, err := h.svc.client.ChunithmDefaultServer.
			Create().
			SetHarukiUserID(userID).
			SetServer(server).
			Save(ctx); err != nil {
			return api.InternalError(c)
		}
	}

	h.svc.ClearDefaultServerCache(ctx, userID)
	return api.JSONResponse(c, fiber.StatusOK, "Default server set")
}

func (h *BindingHandler) DeleteDefaultServer(c fiber.Ctx) error {
	ctx := context.Background()
	userID := api.GetHarukiUserIDFromPath(c)
	if userID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid haruki_user_id")
	}

	count, err := h.svc.client.ChunithmDefaultServer.
		Delete().
		Where(chunithmdefaultserver.HarukiUserIDEQ(userID)).
		Exec(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	if count == 0 {
		return api.JSONResponse(c, fiber.StatusNotFound, "Default server not set")
	}

	h.svc.ClearDefaultServerCache(ctx, userID)
	return api.JSONResponse(c, fiber.StatusOK, "Default server deleted")
}

func (h *BindingHandler) GetBinding(c fiber.Ctx) error {
	ctx := context.Background()
	userID := api.GetHarukiUserIDFromPath(c)
	if userID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid haruki_user_id")
	}
	server := c.Params("server")

	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSBinding)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}

	row, err := h.svc.client.ChunithmBinding.
		Query().
		Where(chunithmbinding.HarukiUserIDEQ(userID), chunithmbinding.ServerEQ(server)).
		First(ctx)
	if err != nil {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrBindingNotFound)
	}

	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", BindingSchema{
		UserID: row.HarukiUserID,
		Server: &row.Server,
		AimeID: &row.AimeID,
	})
}

func (h *BindingHandler) SetBinding(c fiber.Ctx) error {
	ctx := context.Background()
	userID := api.GetHarukiUserIDFromPath(c)
	if userID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid haruki_user_id")
	}
	server := c.Params("server")
	aimeID := c.Params("aime_id")

	row, _ := h.svc.client.ChunithmBinding.
		Query().
		Where(chunithmbinding.HarukiUserIDEQ(userID), chunithmbinding.ServerEQ(server)).
		First(ctx)

	if row != nil {
		if _, err := row.Update().SetAimeID(aimeID).Save(ctx); err != nil {
			return api.InternalError(c)
		}
	} else {
		if _, err := h.svc.client.ChunithmBinding.
			Create().
			SetHarukiUserID(userID).
			SetServer(server).
			SetAimeID(aimeID).
			Save(ctx); err != nil {
			return api.InternalError(c)
		}
	}

	h.svc.ClearBindingCache(ctx, userID, server)
	return api.JSONResponse(c, fiber.StatusOK, "Binding updated")
}

func (h *BindingHandler) DeleteBinding(c fiber.Ctx) error {
	ctx := context.Background()
	userID := api.GetHarukiUserIDFromPath(c)
	if userID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid haruki_user_id")
	}
	server := c.Params("server")
	aimeID := c.Params("aime_id")

	count, err := h.svc.client.ChunithmBinding.
		Delete().
		Where(
			chunithmbinding.HarukiUserIDEQ(userID),
			chunithmbinding.ServerEQ(server),
			chunithmbinding.AimeIDEQ(aimeID),
		).
		Exec(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	if count == 0 {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrBindingNotFound)
	}

	h.svc.ClearBindingCache(ctx, userID, server)
	return api.JSONResponse(c, fiber.StatusOK, "Binding deleted")
}

// ================= Route Registration =================

func registerBindingRoutes(router fiber.Router, client *entchuniMain.Client, redisClient *redis.Client, usersClient *users.Client) {
	svc := NewBindingService(client, redisClient, usersClient)
	h := NewBindingHandler(svc)

	r := router.Group("/user/:haruki_user_id", api.VerifyAPIAuthorization())

	r.Get("/default", h.GetDefaultServer)
	r.Put("/default/:server", h.SetDefaultServer)
	r.Delete("/default", h.DeleteDefaultServer)
	r.Get("/:server", h.GetBinding)
	r.Put("/:server/:aime_id", h.SetBinding)
	r.Delete("/:server/:aime_id", h.DeleteBinding)
}
