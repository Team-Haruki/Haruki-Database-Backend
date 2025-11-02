package pjsk

import (
	"context"
	"fmt"
	"haruki-database/api"
	"haruki-database/config"
	"haruki-database/database/schema/pjsk"
	"haruki-database/database/schema/pjsk/alias"
	"haruki-database/database/schema/pjsk/aliasadmin"
	"haruki-database/database/schema/pjsk/groupalias"
	"haruki-database/database/schema/pjsk/pendingalias"
	"haruki-database/database/schema/pjsk/rejectedalias"
	"haruki-database/utils"
	harukiRedis "haruki-database/utils/redis"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func isAliasAdmin(ctx context.Context, client *pjsk.Client, platform string, imID string) (bool, error) {
	_, err := client.AliasAdmin.
		Query().
		Where(aliasadmin.PlatformEQ(platform), aliasadmin.ImIDEQ(imID)).
		First(ctx)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func requireAliasAdmin(client *pjsk.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		platform := c.Query("platform")
		imID := c.Query("im_id")
		if platform == "" || imID == "" {
			return api.JSONResponse(c, http.StatusBadRequest, "platform and im_id are required")
		}

		ok, err := isAliasAdmin(context.Background(), client, platform, imID)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if !ok {
			return api.JSONResponse(c, http.StatusUnauthorized, "Permission denied")
		}

		return c.Next()
	}
}

// ================= Group Alias Handlers =================
func getGroupAliasToID(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		aliasType := c.Params("alias_type")
		if _, err := utils.ParseAliasType(aliasType); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}
		aliasStr := c.Query("alias")
		platform := c.Params("platform")
		groupID := c.Params("group_id")

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "pjsk-alias")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		rows, err := client.GroupAlias.
			Query().
			Where(
				groupalias.AliasTypeEQ(aliasType),
				groupalias.AliasEQ(aliasStr),
				groupalias.GroupIDEQ(groupID),
				groupalias.PlatformEQ(platform),
			).
			All(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if len(rows) == 0 {
			return api.JSONResponse(c, http.StatusNotFound, "Alias not found")
		}

		ids := make([]int, len(rows))
		for i, r := range rows {
			ids[i] = r.AliasTypeID
		}

		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", AliasToObjectIdResponse{MatchIDs: ids})
	}
}

func getGroupAliasesByID(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		groupID := c.Params("group_id")
		aliasType := c.Params("alias_type")
		if _, err := utils.ParseAliasType(aliasType); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}
		aliasTypeID, _ := strconv.Atoi(c.Params("alias_type_id"))

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "pjsk-alias")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		rows, err := client.GroupAlias.
			Query().
			Where(
				groupalias.PlatformEQ(platform),
				groupalias.GroupIDEQ(groupID),
				groupalias.AliasTypeEQ(aliasType),
				groupalias.AliasTypeIDEQ(aliasTypeID),
			).
			All(ctx)
		if err != nil {
			return api.JSONResponse(c, 500, err.Error())
		}
		if len(rows) == 0 {
			return api.JSONResponse(c, 404, "No aliases found for this group")
		}
		aliases := make([]string, len(rows))
		for i, r := range rows {
			aliases[i] = r.Alias
		}

		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", AllAliasesResponse{Aliases: aliases})
	}
}

func addGroupAlias(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		groupID := c.Params("group_id")
		aliasType := c.Params("alias_type")
		if _, err := utils.ParseAliasType(aliasType); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}
		aliasTypeID, _ := strconv.Atoi(c.Params("alias_type_id"))

		var req AliasRequest
		if err := c.BodyParser(&req); err != nil {
			return api.JSONResponse(c, 400, "Invalid request")
		}

		_, err := client.GroupAlias.
			Create().
			SetPlatform(platform).
			SetGroupID(groupID).
			SetAliasType(aliasType).
			SetAliasTypeID(aliasTypeID).
			SetAlias(req.Alias).
			Save(ctx)
		if err != nil {
			return api.JSONResponse(c, 500, err.Error())
		}

		query := fmt.Sprintf("alias=%s", req.Alias)
		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-alias", fmt.Sprintf("/pjsk/alias/group/%s/%s/%s/%d", platform, groupID, aliasType, aliasTypeID), nil)
		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-alias", fmt.Sprintf("/pjsk/alias/group/%s/%s/%s-id", platform, groupID, aliasType), &query)

		return api.JSONResponse(c, http.StatusOK, "Group alias added")
	}
}

func deleteGroupAlias(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		groupID := c.Params("group_id")
		aliasType := c.Params("alias_type")
		if _, err := utils.ParseAliasType(aliasType); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}
		aliasTypeID, _ := strconv.Atoi(c.Params("alias_type_id"))

		var req AliasRequest
		if err := c.BodyParser(&req); err != nil {
			return api.JSONResponse(c, 400, "Invalid request")
		}

		_, err := client.GroupAlias.
			Delete().
			Where(
				groupalias.PlatformEQ(platform),
				groupalias.GroupIDEQ(groupID),
				groupalias.AliasTypeEQ(aliasType),
				groupalias.AliasTypeIDEQ(aliasTypeID),
				groupalias.AliasEQ(req.Alias),
			).
			Exec(ctx)
		if err != nil {
			return api.JSONResponse(c, 500, err.Error())
		}

		query := fmt.Sprintf("alias=%s", req.Alias)
		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-alias", fmt.Sprintf("/pjsk/alias/group/%s/%s/%s/%d", platform, groupID, aliasType, aliasTypeID), nil)
		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-alias", fmt.Sprintf("/pjsk/alias/group/%s/%s/%s-id", platform, groupID, aliasType), &query)

		return api.JSONResponse(c, http.StatusOK, "Group alias deleted")
	}
}

// ================= Alias Management Handlers =================
func getPendingAliases(client *pjsk.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		rows, err := client.PendingAlias.Query().All(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if len(rows) == 0 {
			return api.JSONResponse(c, http.StatusNotFound, "No pending aliases")
		}
		resp := make([]PendingAlias, len(rows))
		for i, r := range rows {
			resp[i] = PendingAlias{
				ID:          r.ID,
				AliasType:   r.AliasType,
				AliasTypeID: r.AliasTypeID,
				Alias:       r.Alias,
				SubmittedAt: r.SubmittedAt,
				SubmittedBy: r.SubmittedBy,
			}
		}
		return api.JSONResponse(c, http.StatusOK, "ok", resp)
	}
}

func approvePendingAlias(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		pendingID, _ := strconv.Atoi(c.Params("pending_id"))
		row, err := client.PendingAlias.Get(ctx, int64(pendingID))
		if err != nil {
			return api.JSONResponse(c, http.StatusNotFound, "Pending alias not found")
		}
		_, err = client.Alias.
			Create().
			SetAliasType(row.AliasType).
			SetAliasTypeID(row.AliasTypeID).
			SetAlias(row.Alias).
			Save(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		_, err = client.PendingAlias.Delete().Where(pendingalias.IDEQ(int64(pendingID))).Exec(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		query := fmt.Sprintf("alias=%s", row.Alias)
		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-alias", fmt.Sprintf("/pjsk/alias/%s/%d", row.AliasType, row.AliasTypeID), nil)
		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-alias", fmt.Sprintf("/pjsk/alias/%s-id", row.AliasType), &query)
		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-alias", fmt.Sprintf("/pjsk/alias/status/%d", pendingID), nil)

		return api.JSONResponse(c, http.StatusOK, "Alias approved")
	}
}

func rejectPendingAlias(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		pendingID, _ := strconv.Atoi(c.Params("pending_id"))
		platform := c.Query("platform")
		imID := c.Query("im_id")
		row, err := client.PendingAlias.Get(ctx, int64(pendingID))
		if err != nil {
			return api.JSONResponse(c, http.StatusNotFound, "Pending alias not found")
		}
		var req RejectRequest
		if err := c.BodyParser(&req); err != nil {
			return api.JSONResponse(c, 400, "Invalid request")
		}
		_, err = client.RejectedAlias.
			Create().
			SetID(int64(pendingID)).
			SetAliasType(row.AliasType).
			SetAliasTypeID(row.AliasTypeID).
			SetAlias(row.Alias).
			SetReviewedBy(fmt.Sprintf("%s-%s", platform, imID)).
			SetReviewedAt(time.Now()).
			SetReason(req.Reason).
			Save(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		_, err = client.PendingAlias.Delete().Where(pendingalias.IDEQ(int64(pendingID))).Exec(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-alias", fmt.Sprintf("/pjsk/alias/status/%d", pendingID), nil)

		return api.JSONResponse(c, http.StatusOK, "Alias rejected")
	}
}

func getAliasStatus(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		pendingID, _ := strconv.Atoi(c.Params("pending_id"))
		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "pjsk-alias")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		_, err = client.PendingAlias.Get(ctx, int64(pendingID))
		if err == nil {
			return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", fiber.Map{"status": "pending"})
		}

		rejected, err2 := client.RejectedAlias.Query().Where(rejectedalias.IDEQ(int64(pendingID))).First(ctx)
		if err2 == nil {
			return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", fiber.Map{"status": "rejected", "reason": rejected.Reason})
		}
		return api.JSONResponse(c, 404, "Not found")
	}
}

// ================= Global Alias Handlers =================
func getGlobalAliasToID(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		aliasType := c.Params("alias_type")
		if _, err := utils.ParseAliasType(aliasType); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}
		aliasStr := c.Query("alias")

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "pjsk-alias")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		rows, err := client.Alias.Query().
			Where(
				alias.AliasTypeEQ(aliasType),
				alias.AliasEQ(aliasStr),
			).
			All(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if len(rows) == 0 {
			return api.JSONResponse(c, http.StatusNotFound, "Alias not found")
		}
		ids := make([]int, len(rows))
		for i, r := range rows {
			ids[i] = r.AliasTypeID
		}
		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", AliasToObjectIdResponse{MatchIDs: ids})
	}
}

func getGlobalAliasesByID(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		aliasType := c.Params("alias_type")
		if _, err := utils.ParseAliasType(aliasType); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}
		aliasTypeID, _ := strconv.Atoi(c.Params("alias_type_id"))

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "pjsk-alias")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		rows, err := client.Alias.Query().
			Where(
				alias.AliasTypeEQ(aliasType),
				alias.AliasTypeIDEQ(aliasTypeID),
			).
			All(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if len(rows) == 0 {
			return api.JSONResponse(c, http.StatusNotFound, "No aliases found")
		}
		aliases := make([]string, len(rows))
		for i, r := range rows {
			aliases[i] = r.Alias
		}
		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", AllAliasesResponse{Aliases: aliases})
	}
}

func addGlobalAlias(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		aliasType := c.Params("alias_type")
		if _, err := utils.ParseAliasType(aliasType); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}
		aliasTypeID, _ := strconv.Atoi(c.Params("alias_type_id"))
		platform := c.Query("platform")
		imID := c.Query("im_id")

		var req AliasRequest
		if err := c.BodyParser(&req); err != nil {
			return api.JSONResponse(c, 400, "Invalid request")
		}

		isAdmin, err := isAliasAdmin(ctx, client, platform, imID)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		if isAdmin {
			_, err := client.Alias.
				Create().
				SetAliasType(aliasType).
				SetAliasTypeID(aliasTypeID).
				SetAlias(req.Alias).
				Save(ctx)
			if err != nil {
				return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
			}
			query := fmt.Sprintf("alias=%s", req.Alias)
			_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-alias", fmt.Sprintf("/pjsk/alias/%s/%d", aliasType, aliasTypeID), nil)
			_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-alias", fmt.Sprintf("/pjsk/alias/%s-id", aliasType), &query)

			return api.JSONResponse(c, http.StatusOK, "Alias added")
		}

		_, err = client.PendingAlias.
			Create().
			SetAliasType(aliasType).
			SetAliasTypeID(aliasTypeID).
			SetAlias(req.Alias).
			SetSubmittedBy(fmt.Sprintf("%s-%s", platform, imID)).
			SetSubmittedAt(time.Now()).
			Save(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		return api.JSONResponse(c, http.StatusOK, "Alias submitted for approval")
	}
}

func deleteGlobalAlias(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		aliasType := c.Params("alias_type")
		if _, err := utils.ParseAliasType(aliasType); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}
		aliasTypeID, _ := strconv.Atoi(c.Params("alias_type_id"))
		var req AliasRequest
		if err := c.BodyParser(&req); err != nil {
			return api.JSONResponse(c, 400, "Invalid request")
		}
		_, err := client.Alias.
			Delete().
			Where(
				alias.AliasTypeEQ(aliasType),
				alias.AliasTypeIDEQ(aliasTypeID),
				alias.AliasEQ(req.Alias),
			).
			Exec(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		query := fmt.Sprintf("alias=%s", req.Alias)
		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-alias", fmt.Sprintf("/pjsk/alias/%s/%d", aliasType, aliasTypeID), nil)
		_ = harukiRedis.ClearCache(ctx, redisClient, "pjsk-alias", fmt.Sprintf("/pjsk/alias/%s-id", aliasType), &query)

		return api.JSONResponse(c, http.StatusOK, "Alias deleted")
	}
}

func registerAliasRoutes(router fiber.Router, client *pjsk.Client, redisClient *redis.Client) {
	r := router.Group("/alias")

	// ================= Group Alias API =================
	r.Get("/group/:platform/:group_id/:alias_type-id", api.VerifyAPIAuthorization(), getGroupAliasToID(client, redisClient))
	r.Get("/group/:platform/:group_id/:alias_type/:alias_type_id", api.VerifyAPIAuthorization(), getGroupAliasesByID(client, redisClient))
	r.Post("/group/:platform/:group_id/:alias_type/:alias_type_id", api.VerifyAPIAuthorization(), addGroupAlias(client, redisClient))
	r.Delete("/group/:platform/:group_id/:alias_type/:alias_type_id", api.VerifyAPIAuthorization(), deleteGroupAlias(client, redisClient))

	// ================= Global Alias API =================
	// ----------------- Alias Manage API -----------------
	r.Get("/pending", api.VerifyAPIAuthorization(), requireAliasAdmin(client), getPendingAliases(client))
	r.Post("/pending/:pending_id/approve", api.VerifyAPIAuthorization(), requireAliasAdmin(client), approvePendingAlias(client, redisClient))
	r.Post("/pending/:pending_id/reject", api.VerifyAPIAuthorization(), requireAliasAdmin(client), rejectPendingAlias(client, redisClient))
	r.Get("/status/:pending_id", api.VerifyAPIAuthorization(), getAliasStatus(client, redisClient))

	// ----------------- Alias Query API -----------------
	r.Get("/:alias_type-id", getGlobalAliasToID(client, redisClient))
	r.Get("/:alias_type/:alias_type_id", getGlobalAliasesByID(client, redisClient))
	r.Post("/:alias_type/:alias_type_id/add", api.VerifyAPIAuthorization(), addGlobalAlias(client, redisClient))
	r.Delete("/:alias_type/:alias_type_id", api.VerifyAPIAuthorization(), requireAliasAdmin(client), deleteGlobalAlias(client, redisClient))
}
