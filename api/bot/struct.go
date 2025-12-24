package bot

import (
	ent "haruki-database/database/schema/bot"

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

type UserService struct {
	dbClient    *ent.Client
	redisClient *redis.Client
}

type StatisticsService struct {
	client *ent.Client
}

type UserHandler struct {
	svc *UserService
}

type StatisticsHandler struct {
	svc *StatisticsService
}
