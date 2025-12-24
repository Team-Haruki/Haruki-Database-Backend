package censor

import (
	"context"
	"haruki-database/api"
	"haruki-database/utils/censor"

	"github.com/gofiber/fiber/v3"
)

func (h *CensorHandler) CensorName(c fiber.Ctx) error {
	ctx := context.Background()
	imUserID := c.Params("imUserID")
	var req NameRequest
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	ok := h.svc.service.CensorName(ctx, imUserID, req.UserID, req.Name, req.Server)
	msg := censor.ResultNonCompliant
	if ok {
		msg = censor.ResultCompliant
	}
	return api.JSONResponse(c, fiber.StatusOK, string(msg))
}

func (h *CensorHandler) CensorShortBio(c fiber.Ctx) error {
	ctx := context.Background()
	imUserID := c.Params("imUserID")
	var req ShortBioRequest
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	ok := h.svc.service.CensorShortBio(ctx, imUserID, req.UserID, req.Content, req.Server)
	msg := censor.ResultNonCompliant
	if ok {
		msg = censor.ResultCompliant
	}
	return api.JSONResponse(c, fiber.StatusOK, string(msg))
}

func RegisterCensorRoutes(app *fiber.App, service *censor.Service) {
	svc := NewCensorService(service)
	h := NewCensorHandler(svc)

	app.Post("/censor/name/:imUserID", api.VerifyAPIAuthorization(), h.CensorName)
	app.Post("/censor/short-bio/:imUserID", api.VerifyAPIAuthorization(), h.CensorShortBio)
}
