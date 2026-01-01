package chunithm

import (
	entchuniMain "haruki-database/database/schema/chunithm/maindb"
	entchuniMusic "haruki-database/database/schema/chunithm/music"
	"haruki-database/database/schema/users"
	"haruki-database/utils/types"

	"github.com/redis/go-redis/v9"
)

// ================= Type Aliases =================

type AliasToMusicIDResponse = types.AliasToIDResponse
type AllAliasesResponse = types.AliasListResponse
type AliasRequest = types.AliasRequest

type MusicInfoSchema = types.ChunithmMusicInfo
type MusicDifficultySchema = types.ChunithmMusicDifficulty
type ChartDataSchema = types.ChunithmChartData
type MusicBatchItemSchema = types.ChunithmMusicBatchItem

type DefaultServerSchema = types.ChunithmDefaultServer
type BindingSchema = types.ChunithmBinding

type MusicAliasSchema = types.ChunithmMusicAlias

// ================= Cache Namespace Constants =================

const (
	CacheNSAlias   = "hdb:chunithm:alias"
	CacheNSBinding = "hdb:chunithm:binding"
	CacheNSMusic   = "hdb:chunithm:music"
)

// ================= Service Structs =================

type AliasService struct {
	client      *entchuniMain.Client
	redisClient *redis.Client
}

type BindingService struct {
	client      *entchuniMain.Client
	redisClient *redis.Client
	usersClient *users.Client
}

type MusicService struct {
	client      *entchuniMusic.Client
	redisClient *redis.Client
}

// ================= Handler Structs =================

type AliasHandler struct {
	svc *AliasService
}

type BindingHandler struct {
	svc *BindingService
}

type MusicHandler struct {
	svc *MusicService
}
