package bot

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	ent "haruki-database/database/schema/bot"

	"github.com/redis/go-redis/v9"
)

// ================= Service Constructors =================

func NewUserService(dbClient *ent.Client, redisClient *redis.Client) *UserService {
	return &UserService{dbClient: dbClient, redisClient: redisClient}
}

func NewStatisticsService(client *ent.Client) *StatisticsService {
	return &StatisticsService{client: client}
}

// ================= Handler Constructors =================

func NewUserHandler(svc *UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func NewStatisticsHandler(svc *StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{svc: svc}
}

// ================= UserService Methods =================

func (s *UserService) setRedisKey(ctx context.Context, pattern string, id interface{}, value string, ttlMinutes int) error {
	key := fmt.Sprintf(pattern, id)
	return s.redisClient.Set(ctx, key, value, time.Duration(ttlMinutes)*time.Minute).Err()
}

func (s *UserService) getRedisKey(ctx context.Context, pattern string, id interface{}) (string, error) {
	key := fmt.Sprintf(pattern, id)
	return s.redisClient.Get(ctx, key).Result()
}

func (s *UserService) delRedisKey(ctx context.Context, pattern string, id interface{}) error {
	key := fmt.Sprintf(pattern, id)
	return s.redisClient.Del(ctx, key).Err()
}

func (s *UserService) cleanupUserRegistrationKeys(ctx context.Context, userID int64) {
	_ = s.delRedisKey(ctx, RedisKeyVerifyCode, userID)
	_ = s.delRedisKey(ctx, RedisKeyOneTimeToken, userID)
	_ = s.delRedisKey(ctx, RedisKeyVerifyStatus, userID)
}

// ================= Utility Functions =================

func generateVerificationCode(length int) string {
	digits := "0123456789"
	code := make([]byte, length)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		code[i] = digits[n.Int64()]
	}
	return string(code)
}
