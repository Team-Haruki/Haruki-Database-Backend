package api

import (
	"haruki-database/config"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type HarukiAPIResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type HarukiAPIDataResponse[T any] struct {
	HarukiAPIResponse
	Data T `json:"data,omitempty"`
}

func JSONResponse(c *fiber.Ctx, status int, message string, data ...interface{}) error {
	if len(data) > 0 {
		return c.Status(status).JSON(fiber.Map{
			"status":  status,
			"message": message,
			"data":    data[0],
		})
	}
	return c.Status(status).JSON(HarukiAPIResponse{Status: status, Message: message})
}

func VerifyAPIAuthorization() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		userAgent := c.Get("User-Agent")

		if config.Cfg.Backend.AcceptAuthorization != "" && authHeader != config.Cfg.Backend.AcceptAuthorization {
			return JSONResponse(c, http.StatusUnauthorized, "Invalid Authorization header")
		}

		if config.Cfg.Backend.AcceptUserAgent != "" && !strings.Contains(userAgent, config.Cfg.Backend.AcceptUserAgent) {
			return JSONResponse(c, http.StatusForbidden, "Invalid User-Agent")
		}

		return c.Next()
	}
}
