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

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// ================= User Handlers =================

func (h *UserHandler) Register(c fiber.Ctx) error {
	ctx := context.Background()
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	if req.UserID == 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, ErrMissingUserID)
	}
	if req.OneTimeToken == "" {
		return api.JSONResponse(c, fiber.StatusBadRequest, ErrMissingOneTimeToken)
	}
	exists, _ := h.svc.dbClient.User.Query().
		Where(user.OwnerUserIDEQ(req.UserID)).
		Exist(ctx)
	if exists {
		return api.JSONResponse(c, fiber.StatusConflict, ErrBotAlreadyRegistered)
	}
	code := generateVerificationCode(6)
	if err := h.svc.setRedisKey(ctx, RedisKeyVerifyCode, req.UserID, code, VerifyCodeTTLMinutes); err != nil {
		return api.InternalError(c)
	}
	if err := h.svc.setRedisKey(ctx, RedisKeyOneTimeToken, req.UserID, req.OneTimeToken, VerifyCodeTTLMinutes); err != nil {
		return api.InternalError(c)
	}
	return api.JSONResponse(c, fiber.StatusOK,
		fmt.Sprintf("Your verification code is %s, expires in %d minutes.", code, VerifyCodeTTLMinutes))
}

func (h *UserHandler) RegisterVerify(c fiber.Ctx) error {
	ctx := context.Background()
	if c.Get("X-VERIFY") != config.Cfg.HarukiBotDB.RegisterVerifyToken {
		return api.JSONResponse(c, fiber.StatusUnauthorized, api.ErrPermissionDenied)
	}
	var req VerifyRequest
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	stored, err := h.svc.getRedisKey(ctx, RedisKeyVerifyCode, req.UserID)
	if errors.Is(err, redis.Nil) {
		return api.JSONResponse(c, fiber.StatusBadRequest, ErrVerifyCodeNotFound)
	}
	if err != nil {
		return api.InternalError(c)
	}
	if stored != req.VerificationCode {
		return api.JSONResponse(c, fiber.StatusBadRequest, ErrVerifyCodeInvalid)
	}
	_ = h.svc.delRedisKey(ctx, RedisKeyVerifyCode, req.UserID)
	_ = h.svc.setRedisKey(ctx, RedisKeyVerifyStatus, req.UserID, "true", VerifyStatusTTLMinutes)
	return api.JSONResponse(c, fiber.StatusOK, "Successfully verified.")
}

func (h *UserHandler) GetCredential(c fiber.Ctx) error {
	ctx := context.Background()
	userID := fiber.Query[int64](c, "user_id")
	if userID == 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, ErrMissingUserID)
	}
	oneTimeToken := c.Query("one_time_token")
	if oneTimeToken == "" {
		return api.JSONResponse(c, fiber.StatusBadRequest, ErrMissingOneTimeToken)
	}
	storedToken, err := h.svc.getRedisKey(ctx, RedisKeyOneTimeToken, userID)
	if errors.Is(err, redis.Nil) || storedToken != oneTimeToken {
		return api.JSONResponse(c, fiber.StatusBadRequest, ErrInvalidOneTimeToken)
	}
	if err != nil {
		return api.InternalError(c)
	}
	status, _ := h.svc.getRedisKey(ctx, RedisKeyVerifyStatus, userID)
	if status != "true" {
		return api.JSONResponse(c, fiber.StatusBadRequest, ErrNotVerified)
	}
	exists, _ := h.svc.dbClient.User.Query().
		Where(user.OwnerUserIDEQ(userID)).
		Exist(ctx)
	if exists {
		return api.JSONResponse(c, fiber.StatusConflict, ErrBotAlreadyRegistered)
	}
	cred := uuid.NewString()
	botID := generateVerificationCode(8)
	botIDInt, _ := strconv.Atoi(botID)
	_, err = h.svc.dbClient.User.
		Create().
		SetOwnerUserID(userID).
		SetBotID(botIDInt).
		SetCredential(cred).
		Save(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	h.svc.cleanupUserRegistrationKeys(ctx, userID)
	payload := jwt.MapClaims{
		"bot_id":     botID,
		"credential": cred,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, payload).
		SignedString([]byte(config.Cfg.HarukiBotDB.CredentialSignToken))
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
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	decoded, err := jwt.Parse(req.Credential, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(config.Cfg.HarukiBotDB.CredentialSignToken), nil
	})
	if err != nil || !decoded.Valid {
		return api.JSONResponse(c, fiber.StatusBadRequest, ErrInvalidCredential)
	}
	claims, ok := decoded.Claims.(jwt.MapClaims)
	if !ok {
		return api.JSONResponse(c, fiber.StatusBadRequest, ErrInvalidCredential)
	}
	tokenBotID := fmt.Sprintf("%v", claims["bot_id"])
	credential := fmt.Sprintf("%v", claims["credential"])
	if tokenBotID != botID {
		return api.JSONResponse(c, fiber.StatusBadRequest, ErrBotIDMismatch)
	}
	botIDInt, _ := strconv.Atoi(botID)
	exist, err := h.svc.dbClient.User.Query().
		Where(user.BotIDEQ(botIDInt), user.CredentialEQ(credential)).
		Exist(ctx)
	if err != nil || !exist {
		return api.JSONResponse(c, fiber.StatusBadRequest, ErrAuthFailed)
	}
	sessionToken := uuid.NewString()
	sessionPayload := jwt.MapClaims{
		"bot_id":        botID,
		"session_token": sessionToken,
		"exp":           time.Now().Add(time.Duration(SessionTokenTTLMinutes) * time.Minute).Unix(),
	}
	sessionJWT, err := jwt.NewWithClaims(jwt.SigningMethodHS256, sessionPayload).
		SignedString([]byte(config.Cfg.HarukiBotDB.LoginSignToken))
	if err != nil {
		return api.InternalError(c)
	}
	_ = h.svc.setRedisKey(ctx, RedisKeySessionToken, botID, sessionJWT, SessionTokenTTLMinutes)
	return api.JSONResponse(c, fiber.StatusOK, "ok", fiber.Map{
		"session_token": sessionJWT,
	})
}

// ================= Route Registration =================

func registerUserRoutes(app *fiber.App, dbClient *ent.Client, redisClient *redis.Client) {
	svc := NewUserService(dbClient, redisClient)
	h := NewUserHandler(svc)
	r := app.Group("/bot", api.VerifyAPIAuthorization())

	r.Post("/register", h.Register)
	r.Post("/register-verify", h.RegisterVerify)
	r.Get("/get-credential", h.GetCredential)
	r.Put("/:bot_id/auth", h.Auth)
}
