package pjsk

import (
	"haruki-database/database/schema/pjsk"
	"haruki-database/utils/types"

	"github.com/redis/go-redis/v9"
)

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

const (
	aliasParamsKey      = "alias_params"
	groupParamsKey      = "group_params"
	bindingUserIDKey    = "binding_haruki_user_id"
	preferenceUserIDKey = "preference_haruki_user_id"
)

const (
	CacheNSAlias      = "pjsk:alias"
	CacheNSBinding    = "pjsk:binding"
	CacheNSPreference = "pjsk:preference"
)

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

type AliasService struct {
	client      *pjsk.Client
	redisClient *redis.Client
}

type BindingService struct {
	client      *pjsk.Client
	redisClient *redis.Client
}

type PreferenceService struct {
	client      *pjsk.Client
	redisClient *redis.Client
}

type AliasHandler struct {
	svc *AliasService
}

type BindingHandler struct {
	svc *BindingService
}

type PreferenceHandler struct {
	svc *PreferenceService
}
