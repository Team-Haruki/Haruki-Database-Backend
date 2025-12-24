package bot

import (
	"context"
	"errors"
	"fmt"
	"haruki-database/api"
	"haruki-database/config"
	ent "haruki-database/database/schema/bot"
	"haruki-database/database/schema/bot/user"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (h *UserHandler) Register(c fiber.Ctx) error {
	ctx := context.Background()
	var req RegisterRequest
	if err := sonic.Unmarshal(c.Body(), &req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	code := generateVerificationCode(6)
	key := fmt.Sprintf("verify_code:%d", req.UserID)
	if err := h.svc.redisClient.Set(ctx, key, code, 10*time.Minute).Err(); err != nil {
		return api.InternalError(c)
	}
	oneTimeTokenKey := fmt.Sprintf("one_time_token:%d", req.UserID)
	if err := h.svc.redisClient.Set(ctx, oneTimeTokenKey, req.OneTimeToken, 10*time.Minute).Err(); err != nil {
		return api.InternalError(c)
	}
	return api.JSONResponse(c, fiber.StatusOK,
		fmt.Sprintf("Your verification code is %s, expires in 10 minutes.", code))
}

func (h *UserHandler) RegisterVerify(c fiber.Ctx) error {
	ctx := context.Background()
	if c.Get("X-VERIFY") != config.Cfg.HarukiBotDB.RegisterVerifyToken {
		return api.JSONResponse(c, fiber.StatusUnauthorized, api.ErrPermissionDenied)
	}
	var req VerifyRequest
	if err := sonic.Unmarshal(c.Body(), &req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	key := fmt.Sprintf("verify_code:%d", req.UserID)
	stored, err := h.svc.redisClient.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Verification code not found")
	}
	if err != nil {
		return api.InternalError(c)
	}
	if stored != req.VerificationCode {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Verification code is invalid")
	}
	h.svc.redisClient.Del(ctx, key)
	h.svc.redisClient.Set(ctx, fmt.Sprintf("verify_status:%d", req.UserID), "true", 10*time.Minute)
	return api.JSONResponse(c, fiber.StatusOK, "Successfully verified.")
}

func (h *UserHandler) GetCredential(c fiber.Ctx) error {
	ctx := context.Background()
	userID := fiber.Query[int](c, "user_id")
	if userID == 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Missing user_id")
	}
	oneTimeToken := c.Query("one_time_token")
	if oneTimeToken == "" {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Missing one_time_token")
	}
	storedToken, err := h.svc.redisClient.Get(ctx, fmt.Sprintf("one_time_token:%d", userID)).Result()
	if errors.Is(err, redis.Nil) || storedToken != oneTimeToken {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid or missing one-time token")
	}
	if err != nil {
		return api.InternalError(c)
	}
	status, _ := h.svc.redisClient.Get(ctx, fmt.Sprintf("verify_status:%d", userID)).Result()
	if status != "true" {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Not verified")
	}
	cred := uuid.NewString()
	botID := generateVerificationCode(8)
	botIDInt, _ := strconv.Atoi(botID)
	_, err = h.svc.dbClient.User.
		Create().
		SetOwnerUserID(int64(userID)).
		SetBotID(botIDInt).
		SetCredential(cred).
		Save(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	payload := jwt.MapClaims{
		"bot_id":     botID,
		"credential": cred,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString([]byte(config.Cfg.HarukiBotDB.CredentialSignToken))
	if err != nil {
		return api.InternalError(c)
	}
	return api.JSONResponse(c, fiber.StatusOK, "ok", fiber.Map{
		"bot_id":     botID,
		"credential": token,
	})
}

func (h *UserHandler) Auth(c fiber.Ctx) error {
	ctx := context.Background()
	botID := c.Params("bot_id")
	var req AuthRequest
	if err := sonic.Unmarshal(c.Body(), &req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	decoded, err := jwt.Parse(req.Credential, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Cfg.HarukiBotDB.CredentialSignToken), nil
	})
	if err != nil || !decoded.Valid {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid credential")
	}
	claims := decoded.Claims.(jwt.MapClaims)
	tokenBotID := fmt.Sprintf("%v", claims["bot_id"])
	credential := fmt.Sprintf("%v", claims["credential"])
	if tokenBotID != botID {
		return api.JSONResponse(c, fiber.StatusBadRequest, "bot_id mismatch")
	}
	botIDInt, _ := strconv.Atoi(botID)
	exist, err := h.svc.dbClient.User.
		Query().
		Where(user.BotIDEQ(botIDInt), user.CredentialEQ(credential)).
		Exist(ctx)
	if err != nil || !exist {
		return api.JSONResponse(c, fiber.StatusBadRequest, "Authentication failed")
	}
	sessionToken := uuid.NewString()
	payload := jwt.MapClaims{
		"bot_id":        botID,
		"session_token": sessionToken,
		"exp":           time.Now().Add(30 * time.Minute).Unix(),
	}
	sessionJWT, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString([]byte(config.Cfg.HarukiBotDB.LoginSignToken))
	h.svc.redisClient.Set(ctx, fmt.Sprintf("%s_session_token", botID), sessionJWT, 30*time.Minute)
	return api.JSONResponse(c, fiber.StatusOK, "ok", fiber.Map{
		"session_token": sessionJWT,
	})
}

func registerUserRoutes(app *fiber.App, dbClient *ent.Client, redisClient *redis.Client) {
	svc := NewUserService(dbClient, redisClient)
	h := NewUserHandler(svc)
	r := app.Group("/bot")

	r.Post("/register", h.Register)
	r.Post("/register-verify", h.RegisterVerify)
	r.Get("/get-credential", h.GetCredential)
	r.Put("/:bot_id/auth", h.Auth)
}
