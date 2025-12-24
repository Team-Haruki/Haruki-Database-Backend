package chunithm

import (
	"context"
	"fmt"
	"haruki-database/api"
	entchuniMain "haruki-database/database/schema/chunithm/maindb"
	entchuniMusic "haruki-database/database/schema/chunithm/music"
	harukiRedis "haruki-database/utils/redis"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func NewAliasService(client *entchuniMain.Client, redisClient *redis.Client) *AliasService {
	return &AliasService{client: client, redisClient: redisClient}
}

func NewBindingService(client *entchuniMain.Client, redisClient *redis.Client) *BindingService {
	return &BindingService{client: client, redisClient: redisClient}
}

func NewMusicService(client *entchuniMusic.Client, redisClient *redis.Client) *MusicService {
	return &MusicService{client: client, redisClient: redisClient}
}

func NewAliasHandler(svc *AliasService) *AliasHandler {
	return &AliasHandler{svc: svc}
}

func NewBindingHandler(svc *BindingService) *BindingHandler {
	return &BindingHandler{svc: svc}
}

func NewMusicHandler(svc *MusicService) *MusicHandler {
	return &MusicHandler{svc: svc}
}

func (s *AliasService) ClearCache(ctx context.Context, musicID int, alias string) {
	query := fmt.Sprintf("alias=%s", alias)
	_ = harukiRedis.ClearCache(ctx, s.redisClient, CacheNSAlias, fmt.Sprintf("/chunithm/alias/%d", musicID), nil)
	_ = harukiRedis.ClearCache(ctx, s.redisClient, CacheNSAlias, "/chunithm/alias/music-id", &query)
}

func (s *BindingService) ClearDefaultServerCache(ctx context.Context, userID int) {
	_ = harukiRedis.ClearCache(ctx, s.redisClient, CacheNSBinding, fmt.Sprintf("/chunithm/user/%d/default", userID), nil)
}

func (s *BindingService) ClearBindingCache(ctx context.Context, userID int, server string) {
	_ = harukiRedis.ClearCache(ctx, s.redisClient, CacheNSBinding, fmt.Sprintf("/chunithm/user/%d/%s", userID, server), nil)
}

func parseBindingUserID() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := api.ParseUserID(c)
		if err != nil {
			return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidUserID)
		}
		c.Locals(bindingUserIDKey, userID)
		return c.Next()
	}
}

func getBindingUserID(c fiber.Ctx) int {
	if id, ok := c.Locals(bindingUserIDKey).(int); ok {
		return id
	}
	return 0
}

func extractMusicIDs(rows []*entchuniMain.ChunithmMusicAlias) []int {
	ids := make([]int, len(rows))
	for i, r := range rows {
		ids[i] = r.MusicID
	}
	return ids
}

func extractAliasStrings(rows []*entchuniMain.ChunithmMusicAlias) []string {
	aliases := make([]string, len(rows))
	for i, r := range rows {
		aliases[i] = r.Alias
	}
	return aliases
}
