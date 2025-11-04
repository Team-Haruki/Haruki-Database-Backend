package bot

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"haruki-database/config"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/bytedance/sonic"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"haruki-database/api"
	ent "haruki-database/database/schema/bot"
	"haruki-database/database/schema/bot/user"

	"github.com/redis/go-redis/v9"
)

type RegisterRequest struct {
	UserID       int64  `json:"user_id"`
	OneTimeToken string `json:"one_time_token"`
}

type VerifyRequest struct {
	UserID           int64  `json:"user_id"`
	VerificationCode string `json:"verification_code"`
}

type AuthRequest struct {
	Credential string `json:"credential"`
}

func generateVerificationCode(length int) string {
	digits := "0123456789"
	code := make([]byte, length)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		code[i] = digits[n.Int64()]
	}
	return string(code)
}

func handleRegister(redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		var req RegisterRequest
		if err := sonic.Unmarshal(c.Body(), &req); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "Invalid request body")
		}

		code := generateVerificationCode(6)
		key := fmt.Sprintf("verify_code:%d", req.UserID)
		if err := redisClient.Set(ctx, key, code, 10*time.Minute).Err(); err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, "Failed to set verification code")
		}

		oneTimeTokenKey := fmt.Sprintf("one_time_token:%d", req.UserID)
		if err := redisClient.Set(ctx, oneTimeTokenKey, req.OneTimeToken, 10*time.Minute).Err(); err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, "Failed to set one-time token")
		}

		return api.JSONResponse(c, http.StatusOK,
			fmt.Sprintf("Your verification code is %s, expires in 10 minutes.", code))
	}
}

func handleRegisterVerify(redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		if c.Get("X-VERIFY") != config.Cfg.HarukiBotDB.RegisterVerifyToken {
			return api.JSONResponse(c, http.StatusUnauthorized, "Access Denied.")
		}

		var req VerifyRequest
		if err := sonic.Unmarshal(c.Body(), &req); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "Invalid request body")
		}

		key := fmt.Sprintf("verify_code:%d", req.UserID)
		stored, err := redisClient.Get(ctx, key).Result()
		if errors.Is(err, redis.Nil) {
			return api.JSONResponse(c, http.StatusBadRequest, "Verification code not found")
		}
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, "Redis error")
		}

		if stored != req.VerificationCode {
			return api.JSONResponse(c, http.StatusBadRequest, "Verification code is invalid")
		}

		redisClient.Del(ctx, key)
		redisClient.Set(ctx, fmt.Sprintf("verify_status:%d", req.UserID), "true", 10*time.Minute)

		return api.JSONResponse(c, http.StatusOK, "Successfully verified.")
	}
}

func handleGetCredential(dbClient *ent.Client, redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		userID := fiber.Query[int](c, "user_id")
		if userID == 0 {
			return api.JSONResponse(c, http.StatusBadRequest, "Missing user_id")
		}

		oneTimeToken := c.Query("one_time_token")
		if oneTimeToken == "" {
			return api.JSONResponse(c, http.StatusBadRequest, "Missing one_time_token")
		}

		storedToken, err := redisClient.Get(ctx, fmt.Sprintf("one_time_token:%d", userID)).Result()
		if errors.Is(err, redis.Nil) || storedToken != oneTimeToken {
			return api.JSONResponse(c, http.StatusBadRequest, "Invalid or missing one-time token")
		}
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, "Redis error")
		}

		status, _ := redisClient.Get(ctx, fmt.Sprintf("verify_status:%d", userID)).Result()
		if status != "true" {
			return api.JSONResponse(c, http.StatusBadRequest, "Not verified")
		}

		cred := uuid.NewString()
		botID := generateVerificationCode(8)
		botIDInt, _ := strconv.Atoi(botID)

		_, err = dbClient.User.
			Create().
			SetOwnerUserID(int64(userID)).
			SetBotID(botIDInt).
			SetCredential(cred).
			Save(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, "DB insert failed")
		}

		payload := jwt.MapClaims{
			"bot_id":     botID,
			"credential": cred,
		}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString([]byte(config.Cfg.HarukiBotDB.CredentialSignToken))
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, "JWT sign failed")
		}

		return api.JSONResponse(c, http.StatusOK, "ok", fiber.Map{
			"bot_id":     botID,
			"credential": token,
		})
	}
}

func handleAuth(dbClient *ent.Client, redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		botID := c.Params("bot_id")

		var req AuthRequest
		if err := sonic.Unmarshal(c.Body(), &req); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "Invalid request body")
		}

		decoded, err := jwt.Parse(req.Credential, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Cfg.HarukiBotDB.CredentialSignToken), nil
		})
		if err != nil || !decoded.Valid {
			return api.JSONResponse(c, http.StatusBadRequest, "Invalid credential")
		}

		claims := decoded.Claims.(jwt.MapClaims)
		tokenBotID := fmt.Sprintf("%v", claims["bot_id"])
		credential := fmt.Sprintf("%v", claims["credential"])

		if tokenBotID != botID {
			return api.JSONResponse(c, http.StatusBadRequest, "bot_id mismatch")
		}

		botIDInt, _ := strconv.Atoi(botID)
		exist, err := dbClient.User.
			Query().
			Where(user.BotIDEQ(botIDInt), user.CredentialEQ(credential)).
			Exist(ctx)
		if err != nil || !exist {
			return api.JSONResponse(c, http.StatusBadRequest, "Authentication failed")
		}

		sessionToken := uuid.NewString()
		payload := jwt.MapClaims{
			"bot_id":        botID,
			"session_token": sessionToken,
			"exp":           time.Now().Add(30 * time.Minute).Unix(),
		}
		sessionJWT, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString([]byte(config.Cfg.HarukiBotDB.LoginSignToken))

		redisClient.Set(ctx, fmt.Sprintf("%s_session_token", botID), sessionJWT, 30*time.Minute)

		return api.JSONResponse(c, http.StatusOK, "ok", fiber.Map{
			"session_token": sessionJWT,
		})
	}
}

func registerUserRoutes(app *fiber.App, dbClient *ent.Client, redisClient *redis.Client) {
	r := app.Group("/bot")

	r.Post("/register", handleRegister(redisClient))
	r.Post("/register-verify", handleRegisterVerify(redisClient))
	r.Get("/get-credential", handleGetCredential(dbClient, redisClient))
	r.Put("/:bot_id/auth", handleAuth(dbClient, redisClient))
}
