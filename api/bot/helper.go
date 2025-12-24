package bot

import (
	"crypto/rand"
	"math/big"

	ent "haruki-database/database/schema/bot"

	"github.com/redis/go-redis/v9"
)

func NewUserService(dbClient *ent.Client, redisClient *redis.Client) *UserService {
	return &UserService{dbClient: dbClient, redisClient: redisClient}
}

func NewStatisticsService(client *ent.Client) *StatisticsService {
	return &StatisticsService{client: client}
}

func NewUserHandler(svc *UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func NewStatisticsHandler(svc *StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{svc: svc}
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
