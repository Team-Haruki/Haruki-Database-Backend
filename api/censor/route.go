package censor

import (
	"context"
	"haruki-database/api"
	"haruki-database/database/schema/users"
	"haruki-database/utils/censor"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func (h *CensorHandler) CensorName(c fiber.Ctx) error {
	ctx := context.Background()
	var req NameRequest
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}

	if req.HarukiUserID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid haruki_user_id")
	}

	ok := h.svc.service.CensorName(ctx, req.HarukiUserID, req.UserID, req.Name, req.Server)
	msg := censor.ResultNonCompliant
	if ok {
		msg = censor.ResultCompliant
	}
	return api.JSONResponse(c, fiber.StatusOK, string(msg))
}

func (h *CensorHandler) CensorShortBio(c fiber.Ctx) error {
	ctx := context.Background()
	var req ShortBioRequest
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}

	if req.HarukiUserID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid haruki_user_id")
	}

	ok := h.svc.service.CensorShortBio(ctx, req.HarukiUserID, req.UserID, req.Content, req.Server)
	msg := censor.ResultNonCompliant
	if ok {
		msg = censor.ResultCompliant
	}
	return api.JSONResponse(c, fiber.StatusOK, string(msg))
}

func RegisterCensorRoutes(app *fiber.App, service *censor.Service, usersClient *users.Client, redisClient *redis.Client) {
	svc := NewCensorService(service)
	h := NewCensorHandler(svc, usersClient, redisClient)

	app.Post("/censor/name", api.VerifyAPIAuthorization(), h.CensorName)
	app.Post("/censor/short-bio", api.VerifyAPIAuthorization(), h.CensorShortBio)
}
