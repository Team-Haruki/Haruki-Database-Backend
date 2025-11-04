package censor

import (
	"context"
	"haruki-database/api"
	"haruki-database/utils/censor"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

func nameHandler(service *censor.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		imUserID := c.Params("imUserID")
		type Req struct {
			Server string `json:"server"`
			UserID string `json:"userID"`
			Name   string `json:"name"`
		}
		var req Req
		if err := c.Bind().Body(&req); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "Invalid request")
		}
		ok := service.CensorName(ctx, imUserID, req.UserID, req.Name, req.Server)
		msg := censor.ResultNonCompliant
		if ok {
			msg = censor.ResultCompliant
		}
		return api.JSONResponse(c, http.StatusOK, string(msg))
	}
}

func shortBioHandler(service *censor.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		imUserID := c.Params("imUserID")
		type Req struct {
			Server  string `json:"server"`
			UserID  string `json:"userID"`
			Content string `json:"content"`
		}
		var req Req
		if err := c.Bind().Body(&req); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "Invalid request")
		}
		ok := service.CensorShortBio(ctx, imUserID, req.UserID, req.Content, req.Server)
		msg := censor.ResultNonCompliant
		if ok {
			msg = censor.ResultCompliant
		}
		return api.JSONResponse(c, http.StatusOK, string(msg))
	}
}

func RegisterCensorRoutes(app *fiber.App, service *censor.Service) {
	app.Post("/censor/name/:imUserID", api.VerifyAPIAuthorization(), nameHandler(service))
	app.Post("/censor/short-bio/:imUserID", api.VerifyAPIAuthorization(), shortBioHandler(service))
}
