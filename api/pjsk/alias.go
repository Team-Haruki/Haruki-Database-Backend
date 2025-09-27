package pjsk

import (
	"context"
	"fmt"
	"haruki-database/database/schema/pjsk/aliasadmin"
	"net/http"
	"strconv"
	"time"

	"haruki-database/api"
	"haruki-database/database/schema/pjsk"
	"haruki-database/database/schema/pjsk/alias"
	"haruki-database/database/schema/pjsk/groupalias"
	"haruki-database/database/schema/pjsk/pendingalias"
	"haruki-database/database/schema/pjsk/rejectedalias"
	"haruki-database/utils"

	"github.com/gofiber/fiber/v2"
)

func IsAliasAdmin(ctx context.Context, client *pjsk.Client, platform string, imID string) (bool, error) {
	_, err := client.AliasAdmin.
		Query().
		Where(aliasadmin.PlatformEQ(platform), aliasadmin.ImIDEQ(imID)).
		First(ctx)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func RequireAliasAdmin(client *pjsk.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		platform := c.Query("platform")
		imID := c.Query("im_id")
		if platform == "" || imID == "" {
			return api.JSONResponse(c, http.StatusBadRequest, "platform and im_id are required")
		}

		ok, err := IsAliasAdmin(context.Background(), client, platform, imID)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if !ok {
			return api.JSONResponse(c, http.StatusUnauthorized, "Permission denied")
		}

		return c.Next()
	}
}

func RegisterAliasRoutes(router fiber.Router, client *pjsk.Client) {
	r := router.Group("/alias")

	// ================= Group Alias API =================
	r.Get("/group/:alias_type-id", api.VerifyAPIAuthorization(), func(c *fiber.Ctx) error {
		ctx := context.Background()
		aliasType := c.Params("alias_type")
		if _, err := utils.ParseAliasType(aliasType); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}
		aliasStr := c.Query("alias")
		platform := c.Query("platform")
		groupID := c.Query("group_id")

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
		return api.JSONResponse(c, http.StatusOK, "ok", AliasToObjectIdResponse{MatchIDs: ids})
	})

	r.Get("/group/:platform/:group_id/:alias_type/:alias_type_id", api.VerifyAPIAuthorization(), func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		groupID := c.Params("group_id")
		aliasType := c.Params("alias_type")
		if _, err := utils.ParseAliasType(aliasType); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}
		aliasTypeID, _ := strconv.Atoi(c.Params("alias_type_id"))

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
		return api.JSONResponse(c, http.StatusOK, "ok", AllAliasesResponse{Aliases: aliases})
	})

	r.Post("/group/:platform/:group_id/:alias_type/:alias_type_id", api.VerifyAPIAuthorization(), func(c *fiber.Ctx) error {
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
		return api.JSONResponse(c, http.StatusOK, "Group alias added")
	})

	r.Delete("/group/:platform/:group_id/:alias_type/:alias_type_id", api.VerifyAPIAuthorization(), func(c *fiber.Ctx) error {
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
		return api.JSONResponse(c, http.StatusOK, "Group alias deleted")
	})

	// ================= Global Alias API =================
	// ----------------- Alias Manage API -----------------
	r.Get("/pending", api.VerifyAPIAuthorization(), RequireAliasAdmin(client), func(c *fiber.Ctx) error {
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
	})

	r.Post("/pending/:pending_id/approve", api.VerifyAPIAuthorization(), RequireAliasAdmin(client), func(c *fiber.Ctx) error {
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
		return api.JSONResponse(c, http.StatusOK, "Alias approved")
	})

	r.Post("/pending/:pending_id/reject", api.VerifyAPIAuthorization(), RequireAliasAdmin(client), func(c *fiber.Ctx) error {
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
			SetAliasType(row.AliasType).
			SetAliasTypeID(row.AliasTypeID).
			SetAlias(row.Alias).
			SetReviewedBy(fmt.Sprintf("%s-%s", platform, imID)).
			SetReason(req.Reason).
			Save(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		_, err = client.PendingAlias.Delete().Where(pendingalias.IDEQ(int64(pendingID))).Exec(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		return api.JSONResponse(c, http.StatusOK, "Alias rejected")
	})

	r.Get("/status/:pending_id", api.VerifyAPIAuthorization(), func(c *fiber.Ctx) error {
		ctx := context.Background()
		pendingID, _ := strconv.Atoi(c.Params("pending_id"))
		_, err := client.PendingAlias.Get(ctx, int64(pendingID))
		if err == nil {
			return api.JSONResponse(c, http.StatusOK, "ok", fiber.Map{"status": "pending"})
		}
		rejected, err2 := client.RejectedAlias.Query().Where(rejectedalias.IDEQ(int64(pendingID))).First(ctx)
		if err2 == nil {
			return api.JSONResponse(c, http.StatusOK, "ok", fiber.Map{"status": "rejected", "reason": rejected.Reason})
		}
		return api.JSONResponse(c, 404, "Not found")
	})

	// ----------------- Alias Query API -----------------
	r.Get("/:alias_type-id", func(c *fiber.Ctx) error {
		ctx := context.Background()
		aliasType := c.Params("alias_type")
		if _, err := utils.ParseAliasType(aliasType); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}
		aliasStr := c.Query("alias")
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
		return api.JSONResponse(c, http.StatusOK, "ok", AliasToObjectIdResponse{MatchIDs: ids})
	})

	r.Get("/:alias_type/:alias_type_id", func(c *fiber.Ctx) error {
		ctx := context.Background()
		aliasType := c.Params("alias_type")
		if _, err := utils.ParseAliasType(aliasType); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}
		aliasTypeID, _ := strconv.Atoi(c.Params("alias_type_id"))
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
		return api.JSONResponse(c, http.StatusOK, "ok", AllAliasesResponse{Aliases: aliases})
	})

	r.Post("/:alias_type/:alias_type_id/add", api.VerifyAPIAuthorization(), func(c *fiber.Ctx) error {
		ctx := context.Background()
		aliasType := c.Params("alias_type")
		if _, err := utils.ParseAliasType(aliasType); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}
		aliasTypeID, _ := strconv.Atoi(c.Params("alias_type_id"))
		platform := c.Query("platform")
		imID := c.Query("im_id")

		var req struct {
			Alias     string `json:"alias"`
			Submitter string `json:"submitter"`
		}
		if err := c.BodyParser(&req); err != nil {
			return api.JSONResponse(c, 400, "Invalid request")
		}

		isAdmin, err := IsAliasAdmin(ctx, client, platform, imID)
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
			return api.JSONResponse(c, http.StatusOK, "Alias added")
		} else {
			_, err := client.PendingAlias.
				Create().
				SetAliasType(aliasType).
				SetAliasTypeID(aliasTypeID).
				SetAlias(req.Alias).
				SetSubmittedBy(imID).
				SetSubmittedAt(time.Now()).
				Save(ctx)
			if err != nil {
				return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
			}
			return api.JSONResponse(c, http.StatusOK, "Alias submitted for approval")
		}
	})

	r.Delete("/:alias_type/:alias_type_id", api.VerifyAPIAuthorization(), RequireAliasAdmin(client), func(c *fiber.Ctx) error {
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
		return api.JSONResponse(c, http.StatusOK, "Alias deleted")
	})
}
