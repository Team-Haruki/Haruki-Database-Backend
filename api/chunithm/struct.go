package chunithm

import (
	entchuniMain "haruki-database/database/schema/chunithm/maindb"
	entchuniMusic "haruki-database/database/schema/chunithm/music"
	"haruki-database/utils/types"

	"github.com/redis/go-redis/v9"
)

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

const (
	CacheNSAlias   = "chunithm:alias"
	CacheNSBinding = "chunithm:binding"
	CacheNSMusic   = "chunithm:music"
)

const (
	bindingUserIDKey = "chunithm_user_id"
)

type AliasService struct {
	client      *entchuniMain.Client
	redisClient *redis.Client
}

type BindingService struct {
	client      *entchuniMain.Client
	redisClient *redis.Client
}

type MusicService struct {
	client      *entchuniMusic.Client
	redisClient *redis.Client
}

type AliasHandler struct {
	svc *AliasService
}

type BindingHandler struct {
	svc *BindingService
}

type MusicHandler struct {
	svc *MusicService
}
