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

func ClearAllCacheForPath(ctx context.Context, redisClient *redis.Client, namespace, path string) error {
	pattern := fmt.Sprintf("%s:%s:query=*", namespace, path)
	var cursor uint64 = 0
	var keys []string
	for {
		var scannedKeys []string
		var err error
		scannedKeys, cursor, err = redisClient.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan redis keys: %w", err)
		}
		keys = append(keys, scannedKeys...)
		if cursor == 0 {
			break
		}
	}
	if len(keys) == 0 {
		return nil
	}
	if err := redisClient.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete redis keys: %w", err)
	}
	return nil
}
