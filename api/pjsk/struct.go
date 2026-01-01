package pjsk

import (
	"haruki-database/database/schema/pjsk"
	"haruki-database/database/schema/users"
	"haruki-database/utils/types"

	"github.com/redis/go-redis/v9"
)

// ================= Type Aliases =================

type AliasToObjectIdResponse = types.AliasToIDResponse
type AllAliasesResponse = types.AliasListResponse
type AliasRequest = types.AliasRequest
type RejectRequest = types.RejectRequest
type PendingAlias = types.PJSKPendingAlias

type UserPreferenceSchema = types.PJSKPreference
type UserPreferenceResponse = types.PJSKPreferenceResponse

type BindingSchema = types.PJSKBinding
type BindingResponse = types.PJSKBindingResponse
type AddBindingSuccessResponse = types.PJSKAddBindingResponse

// ================= Cache Namespace Constants =================

const (
	CacheNSAlias      = "hdb:pjsk:alias"
	CacheNSBinding    = "hdb:pjsk:binding"
	CacheNSPreference = "hdb:pjsk:preference"
)

// ================= Parameter Structs =================

type AliasParams struct {
	AliasType   string
	AliasTypeID int
	AliasStr    string
}

type GroupAliasParams struct {
	AliasParams
	Platform string
	GroupID  string
}

// ================= Service Structs =================

type AliasService struct {
	client      *pjsk.Client
	redisClient *redis.Client
	usersClient *users.Client
}

type BindingService struct {
	client      *pjsk.Client
	redisClient *redis.Client
	usersClient *users.Client
}

type PreferenceService struct {
	client      *pjsk.Client
	redisClient *redis.Client
	usersClient *users.Client
}

// ================= Handler Structs =================

type AliasHandler struct {
	svc *AliasService
}

type BindingHandler struct {
	svc *BindingService
}

type PreferenceHandler struct {
	svc *PreferenceService
}
