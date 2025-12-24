package pjsk

import (
	"context"
	"haruki-database/api"
	"haruki-database/config"
	"haruki-database/database/schema/pjsk"
	"haruki-database/database/schema/pjsk/alias"
	"haruki-database/database/schema/pjsk/groupalias"
	"haruki-database/database/schema/pjsk/pendingalias"
	"haruki-database/database/schema/pjsk/rejectedalias"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func (h *AliasHandler) GetGroupAliasToID(c fiber.Ctx) error {
	ctx := context.Background()
	params := getGroupAliasParams(c)
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSAlias)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	rows, err := h.svc.client.GroupAlias.
		Query().
		Where(
			groupalias.PlatformEQ(params.Platform),
			groupalias.GroupIDEQ(params.GroupID),
			groupalias.AliasTypeEQ(params.AliasType),
			groupalias.AliasEQ(params.AliasStr),
		).
		All(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	if len(rows) == 0 {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrAliasNotFound)
	}
	ids := extractAliasTypeIDs(rows)
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", AliasToObjectIdResponse{MatchIDs: ids})
}

func (h *AliasHandler) GetGroupAliasesByID(c fiber.Ctx) error {
	ctx := context.Background()
	params := getGroupAliasParams(c)
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSAlias)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	rows, err := h.svc.client.GroupAlias.
		Query().
		Where(
			groupalias.PlatformEQ(params.Platform),
			groupalias.GroupIDEQ(params.GroupID),
			groupalias.AliasTypeEQ(params.AliasType),
			groupalias.AliasTypeIDEQ(params.AliasTypeID),
		).
		All(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	if len(rows) == 0 {
		return api.JSONResponse(c, fiber.StatusNotFound, "No aliases found for this group")
	}
	aliases := extractGroupAliasStrings(rows)
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", AllAliasesResponse{Aliases: aliases})
}

func (h *AliasHandler) AddGroupAlias(c fiber.Ctx) error {
	ctx := context.Background()
	params := getGroupAliasParams(c)
	var req AliasRequest
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	if !api.ValidateAlias(req.Alias) {
		return api.JSONResponse(c, fiber.StatusBadRequest, "invalid alias")
	}
	_, err := h.svc.client.GroupAlias.
		Create().
		SetPlatform(params.Platform).
		SetGroupID(params.GroupID).
		SetAliasType(params.AliasType).
		SetAliasTypeID(params.AliasTypeID).
		SetAlias(req.Alias).
		Save(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	h.svc.ClearGroupCache(ctx, params.Platform, params.GroupID, params.AliasType, params.AliasTypeID, req.Alias)
	return api.JSONResponse(c, fiber.StatusOK, "Group alias added")
}

func (h *AliasHandler) DeleteGroupAlias(c fiber.Ctx) error {
	ctx := context.Background()
	params := getGroupAliasParams(c)
	var req AliasRequest
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	_, err := h.svc.client.GroupAlias.
		Delete().
		Where(
			groupalias.PlatformEQ(params.Platform),
			groupalias.GroupIDEQ(params.GroupID),
			groupalias.AliasTypeEQ(params.AliasType),
			groupalias.AliasTypeIDEQ(params.AliasTypeID),
			groupalias.AliasEQ(req.Alias),
		).
		Exec(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	h.svc.ClearGroupCache(ctx, params.Platform, params.GroupID, params.AliasType, params.AliasTypeID, req.Alias)
	return api.JSONResponse(c, fiber.StatusOK, "Group alias deleted")
}

func (h *AliasHandler) GetPendingAliases(c fiber.Ctx) error {
	ctx := context.Background()
	rows, err := h.svc.client.PendingAlias.Query().All(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	if len(rows) == 0 {
		return api.JSONResponse(c, fiber.StatusNotFound, "No pending aliases")
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
	return api.JSONResponse(c, fiber.StatusOK, "ok", resp)
}

func (h *AliasHandler) ApprovePendingAlias(c fiber.Ctx) error {
	ctx := context.Background()
	pendingID := fiber.Params[int64](c, "pending_id", 0)
	if pendingID == 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "invalid pending_id")
	}
	row, err := h.svc.client.PendingAlias.Get(ctx, pendingID)
	if err != nil {
		return api.JSONResponse(c, fiber.StatusNotFound, "Pending alias not found")
	}
	if _, err = h.svc.client.Alias.
		Create().
		SetAliasType(row.AliasType).
		SetAliasTypeID(row.AliasTypeID).
		SetAlias(row.Alias).
		Save(ctx); err != nil {
		return api.InternalError(c)
	}
	if _, err = h.svc.client.PendingAlias.Delete().Where(pendingalias.IDEQ(pendingID)).Exec(ctx); err != nil {
		return api.InternalError(c)
	}
	h.svc.ClearGlobalCache(ctx, row.AliasType, row.AliasTypeID, row.Alias)
	h.svc.ClearStatusCache(ctx, pendingID)
	return api.JSONResponse(c, fiber.StatusOK, "Alias approved")
}

func (h *AliasHandler) RejectPendingAlias(c fiber.Ctx) error {
	ctx := context.Background()
	pendingID := fiber.Params[int64](c, "pending_id", 0)
	if pendingID == 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "invalid pending_id")
	}
	harukiUserID := fiber.Query[int](c, "haruki_user_id", 0)
	row, err := h.svc.client.PendingAlias.Get(ctx, pendingID)
	if err != nil {
		return api.JSONResponse(c, fiber.StatusNotFound, "Pending alias not found")
	}
	var req RejectRequest
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	if !api.ValidateStringLength(req.Reason, api.MaxReasonLength) {
		return api.JSONResponse(c, fiber.StatusBadRequest, "reason too long")
	}
	if _, err = h.svc.client.RejectedAlias.
		Create().
		SetID(pendingID).
		SetAliasType(row.AliasType).
		SetAliasTypeID(row.AliasTypeID).
		SetAlias(row.Alias).
		SetReviewedBy(strconv.Itoa(harukiUserID)).
		SetReviewedAt(time.Now()).
		SetReason(req.Reason).
		Save(ctx); err != nil {
		return api.InternalError(c)
	}
	if _, err = h.svc.client.PendingAlias.Delete().Where(pendingalias.IDEQ(pendingID)).Exec(ctx); err != nil {
		return api.InternalError(c)
	}
	h.svc.ClearStatusCache(ctx, pendingID)
	return api.JSONResponse(c, fiber.StatusOK, "Alias rejected")
}

func (h *AliasHandler) GetAliasStatus(c fiber.Ctx) error {
	ctx := context.Background()
	pendingID := fiber.Params[int64](c, "pending_id", 0)
	if pendingID == 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "invalid pending_id")
	}
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSAlias)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	if _, err = h.svc.client.PendingAlias.Get(ctx, pendingID); err == nil {
		return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", fiber.Map{"status": "pending"})
	}
	if rejected, err := h.svc.client.RejectedAlias.Query().Where(rejectedalias.IDEQ(pendingID)).First(ctx); err == nil {
		return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", fiber.Map{"status": "rejected", "reason": rejected.Reason})
	}
	return api.JSONResponse(c, fiber.StatusNotFound, "Not found")
}

func (h *AliasHandler) GetGlobalAliasToID(c fiber.Ctx) error {
	ctx := context.Background()
	params := getAliasParams(c)
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSAlias)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	rows, err := h.svc.client.Alias.Query().
		Where(
			alias.AliasTypeEQ(params.AliasType),
			alias.AliasEQ(params.AliasStr),
		).
		All(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	if len(rows) == 0 {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrAliasNotFound)
	}
	ids := extractGlobalAliasTypeIDs(rows)
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", AliasToObjectIdResponse{MatchIDs: ids})
}

func (h *AliasHandler) GetGlobalAliasesByID(c fiber.Ctx) error {
	ctx := context.Background()
	params := getAliasParams(c)
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSAlias)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	rows, err := h.svc.client.Alias.Query().
		Where(
			alias.AliasTypeEQ(params.AliasType),
			alias.AliasTypeIDEQ(params.AliasTypeID),
		).
		All(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	if len(rows) == 0 {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrAliasNotFound)
	}
	aliases := extractGlobalAliasStrings(rows)
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", AllAliasesResponse{Aliases: aliases})
}

func (h *AliasHandler) AddGlobalAlias(c fiber.Ctx) error {
	ctx := context.Background()
	params := getAliasParams(c)
	harukiUserID := fiber.Query[int](c, "haruki_user_id", 0)
	if harukiUserID == 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidHarukiUserID)
	}
	var req AliasRequest
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	if !api.ValidateAlias(req.Alias) {
		return api.JSONResponse(c, fiber.StatusBadRequest, "invalid alias")
	}
	isAdmin, err := h.svc.IsAdmin(ctx, harukiUserID)
	if err != nil {
		return api.InternalError(c)
	}
	if isAdmin {
		if _, err := h.svc.client.Alias.
			Create().
			SetAliasType(params.AliasType).
			SetAliasTypeID(params.AliasTypeID).
			SetAlias(req.Alias).
			Save(ctx); err != nil {
			return api.InternalError(c)
		}
		h.svc.ClearGlobalCache(ctx, params.AliasType, params.AliasTypeID, req.Alias)
		return api.JSONResponse(c, fiber.StatusOK, "Alias added")
	}
	if _, err = h.svc.client.PendingAlias.
		Create().
		SetAliasType(params.AliasType).
		SetAliasTypeID(params.AliasTypeID).
		SetAlias(req.Alias).
		SetSubmittedBy(strconv.Itoa(harukiUserID)).
		SetSubmittedAt(time.Now()).
		Save(ctx); err != nil {
		return api.InternalError(c)
	}
	return api.JSONResponse(c, fiber.StatusOK, "Alias submitted for approval")
}

func (h *AliasHandler) DeleteGlobalAlias(c fiber.Ctx) error {
	ctx := context.Background()
	params := getAliasParams(c)
	var req AliasRequest
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	if _, err := h.svc.client.Alias.
		Delete().
		Where(
			alias.AliasTypeEQ(params.AliasType),
			alias.AliasTypeIDEQ(params.AliasTypeID),
			alias.AliasEQ(req.Alias),
		).
		Exec(ctx); err != nil {
		return api.InternalError(c)
	}
	h.svc.ClearGlobalCache(ctx, params.AliasType, params.AliasTypeID, req.Alias)
	return api.JSONResponse(c, fiber.StatusOK, "Alias deleted")
}

func registerAliasRoutes(router fiber.Router, client *pjsk.Client, redisClient *redis.Client) {
	svc := NewAliasService(client, redisClient)
	h := NewAliasHandler(svc)
	r := router.Group("/alias")

	groupRoutes := r.Group("/group/:platform/:group_id/:alias_type")
	groupRoutes.Get("/by-alias",
		api.VerifyAPIAuthorization(),
		parseGroupAliasParams(false, true),
		h.GetGroupAliasToID)
	groupRoutes.Get("/:alias_type_id",
		api.VerifyAPIAuthorization(),
		parseGroupAliasParams(true, false),
		h.GetGroupAliasesByID)
	groupRoutes.Post("/:alias_type_id",
		api.VerifyAPIAuthorization(),
		parseGroupAliasParams(true, false),
		h.AddGroupAlias)
	groupRoutes.Delete("/:alias_type_id",
		api.VerifyAPIAuthorization(),
		parseGroupAliasParams(true, false),
		h.DeleteGroupAlias)
	r.Get("/pending",
		api.VerifyAPIAuthorization(),
		requireAliasAdmin(svc),
		h.GetPendingAliases)
	r.Post("/pending/:pending_id/approve",
		api.VerifyAPIAuthorization(),
		requireAliasAdmin(svc),
		h.ApprovePendingAlias)
	r.Post("/pending/:pending_id/reject",
		api.VerifyAPIAuthorization(),
		requireAliasAdmin(svc),
		h.RejectPendingAlias)
	r.Get("/status/:pending_id",
		api.VerifyAPIAuthorization(),
		h.GetAliasStatus)
	r.Get("/:alias_type/by-alias",
		parseAliasParams(false, true),
		h.GetGlobalAliasToID)
	r.Get("/:alias_type/:alias_type_id",
		parseAliasParams(true, false),
		h.GetGlobalAliasesByID)
	r.Post("/:alias_type/:alias_type_id",
		api.VerifyAPIAuthorization(),
		parseAliasParams(true, false),
		h.AddGlobalAlias)
	r.Delete("/:alias_type/:alias_type_id",
		api.VerifyAPIAuthorization(),
		requireAliasAdmin(svc),
		parseAliasParams(true, false),
		h.DeleteGlobalAlias)
}
