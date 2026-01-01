package bot

import (
	ent "haruki-database/database/schema/bot"

	"github.com/redis/go-redis/v9"
)

// ================= Request Types =================

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

// ================= Redis Key Prefixes =================

const (
	RedisKeyVerifyCode   = "hdb:bot:verify_code:%d"
	RedisKeyOneTimeToken = "hdb:bot:one_time_token:%d"
	RedisKeyVerifyStatus = "hdb:bot:verify_status:%d"
	RedisKeySessionToken = "hdb:bot:session:%s"
)

// ================= Cache Settings =================

const (
	VerifyCodeTTLMinutes   = 10
	VerifyStatusTTLMinutes = 10
	SessionTokenTTLMinutes = 30
)

// ================= Error Messages =================

const (
	ErrMissingUserID        = "missing user_id"
	ErrMissingOneTimeToken  = "missing one_time_token"
	ErrInvalidOneTimeToken  = "invalid or expired one-time token"
	ErrVerifyCodeNotFound   = "verification code not found or expired"
	ErrVerifyCodeInvalid    = "verification code is invalid"
	ErrNotVerified          = "not verified"
	ErrInvalidCredential    = "invalid credential"
	ErrBotIDMismatch        = "bot_id mismatch"
	ErrAuthFailed           = "authentication failed"
	ErrBotAlreadyRegistered = "bot already registered for this user"
)

// ================= Service Structs =================

type UserService struct {
	dbClient    *ent.Client
	redisClient *redis.Client
}

type StatisticsService struct {
	client *ent.Client
}

// ================= Handler Structs =================

type UserHandler struct {
	svc *UserService
}

type StatisticsHandler struct {
	svc *StatisticsService
}
