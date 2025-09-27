package redis

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func ClearCache(ctx context.Context, redisClient *redis.Client, namespace, path string, queryParams *string) error {
	queryHash := "none"
	if queryParams != nil {
		sum := md5.Sum([]byte(*queryParams))
		queryHash = hex.EncodeToString(sum[:])
	}
	if err := DeleteCache(ctx, redisClient, fmt.Sprintf("%s:%s:query=%s", namespace, path, queryHash)); err != nil {
		return errors.New(fmt.Sprintf("clear redis cache failed: %v", err))
	}
	return nil
}
