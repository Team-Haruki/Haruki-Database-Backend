package pjsk

import (
	"context"
	"haruki-database/api"
	"haruki-database/config"
	"haruki-database/database/schema/pjsk"
	"haruki-database/database/schema/pjsk/userbinding"
	"haruki-database/database/schema/pjsk/userdefaultbinding"
	"haruki-database/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func (h *BindingHandler) GetBindings(c fiber.Ctx) error {
	ctx := context.Background()
	harukiUserID := getBindingUserID(c)
	server := c.Query("server")
	if server != "" {
		if _, err := utils.ParseBindingServer(server); err != nil {
			return api.JSONResponse(c, fiber.StatusBadRequest, err.Error())
		}
	}
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSBinding)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	q := h.svc.client.UserBinding.Query().Where(userbinding.HarukiUserIDEQ(harukiUserID))
	if server != "" {
		q = q.Where(userbinding.ServerEQ(server))
	}
	rows, err := q.All(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	if len(rows) == 0 {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrBindingNotFound)
	}
	out := make([]BindingSchema, len(rows))
	for i, r := range rows {
		out[i] = BindingSchema{ID: r.ID, HarukiUserID: r.HarukiUserID, Server: r.Server, UserID: r.UserID, Visible: r.Visible}
	}
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", BindingResponse{Bindings: out})
}

func (h *BindingHandler) CreateBinding(c fiber.Ctx) error {
	ctx := context.Background()
	harukiUserID := getBindingUserID(c)
	var body struct {
		Server  string `json:"server"`
		UserID  string `json:"user_id"`
		Visible bool   `json:"visible"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	if serverEnum, _ := utils.ParseDefaultBindingServer(body.Server); serverEnum == utils.DefaultBindingServerDefault {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Unacceptable server param")
	}
	if _, err := utils.ParseBindingServer(body.Server); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, err.Error())
	}
	exists, _ := h.svc.client.UserBinding.Query().
		Where(userbinding.HarukiUserIDEQ(harukiUserID), userbinding.ServerEQ(body.Server), userbinding.UserIDEQ(body.UserID)).
		Exist(ctx)
	if exists {
		return api.JSONResponse(c, fiber.StatusConflict, api.ErrAlreadyExists)
	}
	newBind, err := h.svc.client.UserBinding.Create().
		SetHarukiUserID(harukiUserID).
		SetServer(body.Server).
		SetUserID(body.UserID).
		SetVisible(body.Visible).
		Save(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	h.svc.ClearBindingCache(ctx, harukiUserID)
	return api.JSONResponse(c, fiber.StatusCreated, "ok", AddBindingSuccessResponse{BindingID: newBind.ID})
}

func (h *BindingHandler) GetDefaultBinding(c fiber.Ctx) error {
	ctx := context.Background()
	harukiUserID := getBindingUserID(c)
	server := c.Query("server", "default")
	if _, err := utils.ParseDefaultBindingServer(server); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, err.Error())
	}
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSBinding)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	row, err := h.svc.client.UserDefaultBinding.Query().
		Where(userdefaultbinding.HarukiUserIDEQ(harukiUserID), userdefaultbinding.ServerEQ(server)).
		WithBinding().
		First(ctx)
	if err != nil || row.Edges.Binding == nil {
		msg := "No global default set"
		if server != "default" {
			msg = "No default for server '" + server + "'"
		}
		return api.JSONResponse(c, fiber.StatusNotFound, msg)
	}
	b := row.Edges.Binding
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", BindingResponse{
		Binding: &BindingSchema{ID: b.ID, HarukiUserID: b.HarukiUserID, Server: b.Server, UserID: b.UserID, Visible: b.Visible},
	})
}

func (h *BindingHandler) SetDefaultBinding(c fiber.Ctx) error {
	ctx := context.Background()
	harukiUserID := getBindingUserID(c)
	var body struct {
		Server    string `json:"server"`
		BindingID int    `json:"binding_id"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	dfs, err := utils.ParseDefaultBindingServer(body.Server)
	if err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, err.Error())
	}
	binding, err := h.svc.client.UserBinding.Query().
		Where(userbinding.HarukiUserIDEQ(harukiUserID), userbinding.IDEQ(body.BindingID)).
		First(ctx)
	if err != nil {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrBindingNotFound)
	}
	if dfs != utils.DefaultBindingServerDefault && binding.Server != body.Server {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Binding server mismatch")
	}
	_, _ = h.svc.client.UserDefaultBinding.Delete().
		Where(userdefaultbinding.HarukiUserIDEQ(harukiUserID), userdefaultbinding.ServerEQ(body.Server)).
		Exec(ctx)
	if _, err = h.svc.client.UserDefaultBinding.Create().
		SetHarukiUserID(harukiUserID).
		SetServer(body.Server).
		SetBindingID(body.BindingID).
		Save(ctx); err != nil {
		return api.InternalError(c)
	}
	h.svc.ClearBindingCache(ctx, harukiUserID)
	return api.JSONResponse(c, fiber.StatusOK, "Default binding set for "+body.Server)
}

func (h *BindingHandler) DeleteDefaultBinding(c fiber.Ctx) error {
	ctx := context.Background()
	harukiUserID := getBindingUserID(c)
	var body struct{ Server string }
	if err := c.Bind().Body(&body); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	if _, err := utils.ParseDefaultBindingServer(body.Server); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, err.Error())
	}
	if _, err := h.svc.client.UserDefaultBinding.Delete().
		Where(userdefaultbinding.HarukiUserIDEQ(harukiUserID), userdefaultbinding.ServerEQ(body.Server)).
		Exec(ctx); err != nil {
		return api.InternalError(c)
	}
	h.svc.ClearBindingCache(ctx, harukiUserID)
	return api.JSONResponse(c, fiber.StatusOK, "Default binding deleted for "+body.Server)
}

func (h *BindingHandler) UpdateVisibility(c fiber.Ctx) error {
	ctx := context.Background()
	harukiUserID := getBindingUserID(c)
	bindingID := fiber.Params[int](c, "binding_id", 0)
	var body struct{ Visible bool }
	if err := c.Bind().Body(&body); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	if _, err := h.svc.client.UserBinding.Query().
		Where(userbinding.HarukiUserIDEQ(harukiUserID), userbinding.IDEQ(bindingID)).
		First(ctx); err != nil {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrBindingNotFound)
	}
	if _, err := h.svc.client.UserBinding.UpdateOneID(bindingID).
		SetVisible(body.Visible).
		Save(ctx); err != nil {
		return api.InternalError(c)
	}
	h.svc.ClearBindingCache(ctx, harukiUserID)
	return api.JSONResponse(c, fiber.StatusOK, "Visibility updated")
}

func (h *BindingHandler) DeleteBinding(c fiber.Ctx) error {
	ctx := context.Background()
	harukiUserID := getBindingUserID(c)
	bindingID := fiber.Params[int](c, "binding_id", 0)
	_, _ = h.svc.client.UserDefaultBinding.Delete().
		Where(userdefaultbinding.HarukiUserIDEQ(harukiUserID), userdefaultbinding.BindingIDEQ(bindingID)).
		Exec(ctx)
	if _, err := h.svc.client.UserBinding.Delete().
		Where(userbinding.HarukiUserIDEQ(harukiUserID), userbinding.IDEQ(bindingID)).
		Exec(ctx); err != nil {
		return api.InternalError(c)
	}
	h.svc.ClearBindingCache(ctx, harukiUserID)
	return api.JSONResponse(c, fiber.StatusOK, "Binding deleted")
}

func registerBindingRoutes(router fiber.Router, client *pjsk.Client, redisClient *redis.Client) {
	svc := NewBindingService(client, redisClient)
	h := NewBindingHandler(svc)
	r := router.Group("/user/:haruki_user_id", api.VerifyAPIAuthorization(), parseBindingUserID())

	r.Get("/binding", h.GetBindings)
	r.Post("/binding", h.CreateBinding)
	r.Get("/binding/default", h.GetDefaultBinding)
	r.Put("/binding/default", h.SetDefaultBinding)
	r.Delete("/binding/default", h.DeleteDefaultBinding)
	r.Patch("/binding/:binding_id", h.UpdateVisibility)
	r.Delete("/binding/:binding_id", h.DeleteBinding)
}
