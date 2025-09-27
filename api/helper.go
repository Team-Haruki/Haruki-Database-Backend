package api

import (
	"context"
	"haruki-database/config"
	harukiRedis "haruki-database/utils/redis"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type HarukiAPIResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type HarukiAPIDataResponse[T any] struct {
	HarukiAPIResponse
	Data T `json:"data,omitempty"`
}

func BuildResponseMap(status int, message string, data interface{}) fiber.Map {
	return fiber.Map{
		"status":  status,
		"message": message,
		"data":    data,
	}
}

func JSONResponse(c *fiber.Ctx, status int, message string, data ...interface{}) error {
	resp := BuildResponseMap(status, message, data)
	return c.Status(status).JSON(resp)
}

func CachedJSONResponse(
	ctx context.Context,
	c *fiber.Ctx,
	redisClient *redis.Client,
	ttl time.Duration,
	key string,
	status int,
	message string,
	data interface{},
) error {
	resp := BuildResponseMap(status, message, data)
	if err := harukiRedis.SetCache(ctx, redisClient, key, resp, ttl); err != nil {
	}
	return c.Status(status).JSON(resp)
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

func CacheQuery(ctx context.Context, c *fiber.Ctx, redisClient *redis.Client, namespace string) (*string, error) {
	key := harukiRedis.CacheKeyBuilder(c, namespace)
	var cached map[string]any
	if found, err := harukiRedis.GetCache(ctx, redisClient, key, &cached); err == nil && found {
		return nil, c.Status(http.StatusOK).JSON(cached)
	}
	return &key, nil
}
