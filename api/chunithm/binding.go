package chunithm

import (
	"context"
	"haruki-database/api"
	"haruki-database/config"
	entchuniMain "haruki-database/database/schema/chunithm/maindb"
	"haruki-database/database/schema/chunithm/maindb/chunithmbinding"
	"haruki-database/database/schema/chunithm/maindb/chunithmdefaultserver"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func (h *BindingHandler) GetDefaultServer(c fiber.Ctx) error {
	ctx := context.Background()
	userID := getBindingUserID(c)
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSBinding)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	row, err := h.svc.client.ChunithmDefaultServer.
		Query().
		Where(chunithmdefaultserver.UserIDEQ(userID)).
		First(ctx)
	if err != nil {
		return api.JSONResponse(c, fiber.StatusNotFound, "Default server not set")
	}
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", DefaultServerSchema{
		UserID: row.UserID,
		Server: row.Server,
	})
}

func (h *BindingHandler) SetDefaultServer(c fiber.Ctx) error {
	ctx := context.Background()
	userID := getBindingUserID(c)
	server := c.Params("server")
	row, _ := h.svc.client.ChunithmDefaultServer.
		Query().
		Where(chunithmdefaultserver.UserIDEQ(userID)).
		First(ctx)
	if row != nil {
		if _, err := row.Update().SetServer(server).Save(ctx); err != nil {
			return api.InternalError(c)
		}
	} else {
		if _, err := h.svc.client.ChunithmDefaultServer.
			Create().
			SetUserID(userID).
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
	userID := getBindingUserID(c)
	count, err := h.svc.client.ChunithmDefaultServer.
		Delete().
		Where(chunithmdefaultserver.UserIDEQ(userID)).
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
	userID := getBindingUserID(c)
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
		Where(chunithmbinding.UserIDEQ(userID), chunithmbinding.ServerEQ(server)).
		First(ctx)
	if err != nil {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrBindingNotFound)
	}
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", BindingSchema{
		UserID: row.UserID,
		Server: &row.Server,
		AimeID: &row.AimeID,
	})
}

func (h *BindingHandler) SetBinding(c fiber.Ctx) error {
	ctx := context.Background()
	userID := getBindingUserID(c)
	server := c.Params("server")
	aimeID := c.Params("aime_id")
	row, _ := h.svc.client.ChunithmBinding.
		Query().
		Where(chunithmbinding.UserIDEQ(userID), chunithmbinding.ServerEQ(server)).
		First(ctx)
	if row != nil {
		if _, err := row.Update().SetAimeID(aimeID).Save(ctx); err != nil {
			return api.InternalError(c)
		}
	} else {
		if _, err := h.svc.client.ChunithmBinding.
			Create().
			SetUserID(userID).
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
	userID := getBindingUserID(c)
	server := c.Params("server")
	aimeID := c.Params("aime_id")
	count, err := h.svc.client.ChunithmBinding.
		Delete().
		Where(
			chunithmbinding.UserIDEQ(userID),
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

func registerBindingRoutes(router fiber.Router, client *entchuniMain.Client, redisClient *redis.Client) {
	svc := NewBindingService(client, redisClient)
	h := NewBindingHandler(svc)
	r := router.Group("/user/:user_id", api.VerifyAPIAuthorization(), parseBindingUserID())

	r.Get("/default", h.GetDefaultServer)
	r.Put("/default/:server", h.SetDefaultServer)
	r.Delete("/default", h.DeleteDefaultServer)
	r.Get("/:server", h.GetBinding)
	r.Put("/:server/:aime_id", h.SetBinding)
	r.Delete("/:server/:aime_id", h.DeleteBinding)
}
