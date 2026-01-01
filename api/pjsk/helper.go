package pjsk

import (
	"context"
	"fmt"
	"haruki-database/api"
	"haruki-database/database/schema/pjsk"
	"haruki-database/database/schema/pjsk/aliasadmin"
	"haruki-database/database/schema/users"
	"haruki-database/utils"
	harukiRedis "haruki-database/utils/redis"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

// ================= Context Keys =================

const (
	aliasParamsKey = "alias_params"
	groupParamsKey = "group_params"
)

// ================= Service Constructors =================

func NewAliasService(client *pjsk.Client, redisClient *redis.Client, usersClient *users.Client) *AliasService {
	return &AliasService{client: client, redisClient: redisClient, usersClient: usersClient}
}

func NewBindingService(client *pjsk.Client, redisClient *redis.Client, usersClient *users.Client) *BindingService {
	return &BindingService{client: client, redisClient: redisClient, usersClient: usersClient}
}

func NewPreferenceService(client *pjsk.Client, redisClient *redis.Client, usersClient *users.Client) *PreferenceService {
	return &PreferenceService{client: client, redisClient: redisClient, usersClient: usersClient}
}

// ================= Handler Constructors =================

func NewAliasHandler(svc *AliasService) *AliasHandler {
	return &AliasHandler{svc: svc}
}

func NewBindingHandler(svc *BindingService) *BindingHandler {
	return &BindingHandler{svc: svc}
}

func NewPreferenceHandler(svc *PreferenceService) *PreferenceHandler {
	return &PreferenceHandler{svc: svc}
}

// ================= AliasService Methods =================

func (s *AliasService) IsAdmin(ctx context.Context, harukiUserID int) (bool, error) {
	return s.client.AliasAdmin.Query().
		Where(aliasadmin.HarukiUserIDEQ(harukiUserID)).
		Exist(ctx)
}

func (s *AliasService) ClearGlobalCache(ctx context.Context, aliasType string, aliasTypeID int, aliasStr string) {
	query := fmt.Sprintf("alias=%s", aliasStr)
	_ = harukiRedis.ClearCache(ctx, s.redisClient, CacheNSAlias, fmt.Sprintf("/pjsk/alias/%s/%d", aliasType, aliasTypeID), nil)
	_ = harukiRedis.ClearCache(ctx, s.redisClient, CacheNSAlias, fmt.Sprintf("/pjsk/alias/%s/by-alias", aliasType), &query)
}

func (s *AliasService) ClearGroupCache(ctx context.Context, platform, groupID, aliasType string, aliasTypeID int, aliasStr string) {
	query := fmt.Sprintf("alias=%s", aliasStr)
	_ = harukiRedis.ClearCache(ctx, s.redisClient, CacheNSAlias, fmt.Sprintf("/pjsk/alias/group/%s/%s/%s/%d", platform, groupID, aliasType, aliasTypeID), nil)
	_ = harukiRedis.ClearCache(ctx, s.redisClient, CacheNSAlias, fmt.Sprintf("/pjsk/alias/group/%s/%s/%s/by-alias", platform, groupID, aliasType), &query)
}

func (s *AliasService) ClearStatusCache(ctx context.Context, pendingID int64) {
	_ = harukiRedis.ClearCache(ctx, s.redisClient, CacheNSAlias, fmt.Sprintf("/pjsk/alias/status/%d", pendingID), nil)
}

// ================= BindingService Methods =================

func (s *BindingService) ClearBindingCache(ctx context.Context, harukiUserID int) {
	_ = harukiRedis.ClearAllCacheForPath(ctx, s.redisClient, CacheNSBinding, fmt.Sprintf("/pjsk/user/%d/binding", harukiUserID))
	_ = harukiRedis.ClearAllCacheForPath(ctx, s.redisClient, CacheNSBinding, fmt.Sprintf("/pjsk/user/%d/binding/default", harukiUserID))
}

// ================= PreferenceService Methods =================

func (s *PreferenceService) ClearCache(ctx context.Context, harukiUserID int, option string) {
	_ = harukiRedis.ClearAllCacheForPath(ctx, s.redisClient, CacheNSPreference, fmt.Sprintf("/pjsk/user/%d/preference", harukiUserID))
	if option != "" {
		_ = harukiRedis.ClearAllCacheForPath(ctx, s.redisClient, CacheNSPreference, fmt.Sprintf("/pjsk/user/%d/preference/%s", harukiUserID, option))
	}
}

// ================= Alias Middleware =================

func parseAliasParams(requireID bool, requireAlias bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		params := AliasParams{
			AliasType: c.Params("alias_type"),
		}
		if _, err := utils.ParseAliasType(params.AliasType); err != nil {
			return api.JSONResponse(c, fiber.StatusBadRequest, err.Error())
		}
		if requireID {
			params.AliasTypeID = fiber.Params[int](c, "alias_type_id", -1)
			if params.AliasTypeID < 0 {
				return api.JSONResponse(c, fiber.StatusBadRequest, "invalid alias_type_id")
			}
		}
		if requireAlias {
			params.AliasStr = c.Query("alias")
			if !api.ValidateAlias(params.AliasStr) {
				return api.JSONResponse(c, fiber.StatusBadRequest, "invalid alias")
			}
		}
		c.Locals(aliasParamsKey, &params)
		return c.Next()
	}
}

func parseGroupAliasParams(requireID bool, requireAlias bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		params := GroupAliasParams{
			Platform: c.Params("platform"),
			GroupID:  c.Params("group_id"),
		}
		params.AliasType = c.Params("alias_type")
		if _, err := utils.ParseAliasType(params.AliasType); err != nil {
			return api.JSONResponse(c, fiber.StatusBadRequest, err.Error())
		}
		if requireID {
			params.AliasTypeID = fiber.Params[int](c, "alias_type_id", -1)
			if params.AliasTypeID < 0 {
				return api.JSONResponse(c, fiber.StatusBadRequest, "invalid alias_type_id")
			}
		}
		if requireAlias {
			params.AliasStr = c.Query("alias")
			if !api.ValidateAlias(params.AliasStr) {
				return api.JSONResponse(c, fiber.StatusBadRequest, "invalid alias")
			}
		}
		c.Locals(groupParamsKey, &params)
		return c.Next()
	}
}

func requireAliasAdmin(svc *AliasService) fiber.Handler {
	return func(c fiber.Ctx) error {
		harukiUserID := api.GetHarukiUserIDFromQuery(c)
		if harukiUserID <= 0 {
			return api.JSONResponse(c, fiber.StatusBadRequest, "Invalid or missing haruki_user_id")
		}
		ok, err := svc.IsAdmin(context.Background(), harukiUserID)
		if err != nil {
			return api.InternalError(c)
		}
		if !ok {
			return api.JSONResponse(c, fiber.StatusForbidden, api.ErrPermissionDenied)
		}
		return c.Next()
	}
}

// ================= Context Getters =================

func getAliasParams(c fiber.Ctx) *AliasParams {
	if p, ok := c.Locals(aliasParamsKey).(*AliasParams); ok {
		return p
	}
	return nil
}

func getGroupAliasParams(c fiber.Ctx) *GroupAliasParams {
	if p, ok := c.Locals(groupParamsKey).(*GroupAliasParams); ok {
		return p
	}
	return nil
}

// ================= Extract Helpers =================

func extractAliasTypeIDs(rows []*pjsk.GroupAlias) []int {
	ids := make([]int, len(rows))
	for i, r := range rows {
		ids[i] = r.AliasTypeID
	}
	return ids
}

func extractGroupAliasStrings(rows []*pjsk.GroupAlias) []string {
	aliases := make([]string, len(rows))
	for i, r := range rows {
		aliases[i] = r.Alias
	}
	return aliases
}

func extractGlobalAliasTypeIDs(rows []*pjsk.Alias) []int {
	ids := make([]int, len(rows))
	for i, r := range rows {
		ids[i] = r.AliasTypeID
	}
	return ids
}

func extractGlobalAliasStrings(rows []*pjsk.Alias) []string {
	aliases := make([]string, len(rows))
	for i, r := range rows {
		aliases[i] = r.Alias
	}
	return aliases
}
