package chunithm

import (
	"context"
	"fmt"
	entchuniMain "haruki-database/database/schema/chunithm/maindb"
	entchuniMusic "haruki-database/database/schema/chunithm/music"
	"haruki-database/database/schema/users"
	harukiRedis "haruki-database/utils/redis"

	"github.com/redis/go-redis/v9"
)

// ================= Service Constructors =================

func NewAliasService(client *entchuniMain.Client, redisClient *redis.Client) *AliasService {
	return &AliasService{client: client, redisClient: redisClient}
}

func NewBindingService(client *entchuniMain.Client, redisClient *redis.Client, usersClient *users.Client) *BindingService {
	return &BindingService{client: client, redisClient: redisClient, usersClient: usersClient}
}

func NewMusicService(client *entchuniMusic.Client, redisClient *redis.Client) *MusicService {
	return &MusicService{client: client, redisClient: redisClient}
}

// ================= Handler Constructors =================

func NewAliasHandler(svc *AliasService) *AliasHandler {
	return &AliasHandler{svc: svc}
}

func NewBindingHandler(svc *BindingService) *BindingHandler {
	return &BindingHandler{svc: svc}
}

func NewMusicHandler(svc *MusicService) *MusicHandler {
	return &MusicHandler{svc: svc}
}

// ================= AliasService Methods =================

func (s *AliasService) ClearCache(ctx context.Context, musicID int, alias string) {
	query := fmt.Sprintf("alias=%s", alias)
	_ = harukiRedis.ClearCache(ctx, s.redisClient, CacheNSAlias, fmt.Sprintf("/chunithm/alias/%d", musicID), nil)
	_ = harukiRedis.ClearCache(ctx, s.redisClient, CacheNSAlias, "/chunithm/alias/music-id", &query)
}

// ================= BindingService Methods =================

func (s *BindingService) ClearDefaultServerCache(ctx context.Context, userID int) {
	path := fmt.Sprintf("/chunithm/user/%d/default", userID)
	_ = harukiRedis.ClearCache(ctx, s.redisClient, CacheNSBinding, path, nil)
}

func (s *BindingService) ClearBindingCache(ctx context.Context, userID int, server string) {
	path := fmt.Sprintf("/chunithm/user/%d/%s", userID, server)
	_ = harukiRedis.ClearCache(ctx, s.redisClient, CacheNSBinding, path, nil)
}

// ================= Extract Helpers =================

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
