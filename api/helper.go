package api

import (
	"context"
	"haruki-database/config"
	users "haruki-database/database/schema/users"
	"haruki-database/database/schema/users/user"
	harukiRedis "haruki-database/utils/redis"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func BuildResponseMap(status int, message string, data interface{}) fiber.Map {
	return fiber.Map{
		"status":  status,
		"message": message,
		"data":    data,
	}
}

func JSONResponse(c fiber.Ctx, status int, message string, data ...interface{}) error {
	var resp fiber.Map
	if len(data) > 0 {
		resp = BuildResponseMap(status, message, data[0])
	} else {
		resp = BuildResponseMap(status, message, nil)
	}
	return c.Status(status).JSON(resp)
}

func ErrorResponse(c fiber.Ctx, status int, message string) error {
	return JSONResponse(c, status, message)
}

func InternalError(c fiber.Ctx) error {
	return JSONResponse(c, fiber.StatusInternalServerError, ErrInternalServer)
}

func CachedJSONResponse(
	ctx context.Context,
	c fiber.Ctx,
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
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		userAgent := c.Get("User-Agent")

		if config.Cfg.Backend.AcceptAuthorization != "" && authHeader != config.Cfg.Backend.AcceptAuthorization {
			return JSONResponse(c, fiber.StatusUnauthorized, "Invalid Authorization header")
		}

		if config.Cfg.Backend.AcceptUserAgent != "" && !strings.Contains(userAgent, config.Cfg.Backend.AcceptUserAgent) {
			return JSONResponse(c, fiber.StatusForbidden, "Invalid User-Agent")
		}

		return c.Next()
	}
}

func CacheQuery(ctx context.Context, c fiber.Ctx, redisClient *redis.Client, namespace string) (string, map[string]any, bool, error) {
	key := harukiRedis.CacheKeyBuilder(c, namespace)
	var cached map[string]any
	found, err := harukiRedis.GetCache(ctx, redisClient, key, &cached)
	if err != nil {
		return key, nil, false, err
	}
	if found {
		return key, cached, true, nil
	}
	return key, nil, false, nil
}

func ParseHarukiUserID(c fiber.Ctx) (int, error) {
	harukiUserIDStr := c.Params("haruki_user_id")
	return strconv.Atoi(harukiUserIDStr)
}

func ParseUserID(c fiber.Ctx) (int, error) {
	userIDStr := c.Params("user_id")
	return strconv.Atoi(userIDStr)
}

func RequireUser(usersClient *users.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Query("platform")
		platformUserID := c.Query("platform_user_id")
		if platform == "" || platformUserID == "" {
			return JSONResponse(c, fiber.StatusBadRequest, "platform and platform_user_id are required")
		}
		u, err := usersClient.User.
			Query().
			Where(user.PlatformEQ(platform), user.UserIDEQ(platformUserID)).
			First(ctx)
		if err != nil {
			return JSONResponse(c, fiber.StatusNotFound, ErrUserNotFound)
		}
		if u.BanState {
			return JSONResponse(c, fiber.StatusForbidden, "User is banned: "+u.BanReason)
		}
		c.Locals(UserContextKey, &UserInfo{
			HarukiUserID: u.ID,
			Platform:     u.Platform,
			UserID:       u.UserID,
			BanState:     u.BanState,
			BanReason:    u.BanReason,
		})

		return c.Next()
	}
}

func GetUserFromContext(c fiber.Ctx) *UserInfo {
	if u, ok := c.Locals(UserContextKey).(*UserInfo); ok {
		return u
	}
	return nil
}

func ValidateStringLength(s string, maxLen int) bool {
	return utf8.RuneCountInString(s) <= maxLen
}

func ValidateAlias(alias string) bool {
	if alias == "" {
		return false
	}
	return ValidateStringLength(alias, MaxAliasLength)
}

func ValidateServer(server string) bool {
	if server == "" {
		return false
	}
	return ValidateStringLength(server, MaxServerLength)
}
