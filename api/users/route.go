package users

import (
	"context"
	"haruki-database/api"
	"haruki-database/database/schema/users"
	"haruki-database/database/schema/users/user"

	"github.com/gofiber/fiber/v3"
)

func (h *UserHandler) GetUser(c fiber.Ctx) error {
	ctx := context.Background()
	platform := c.Query("platform")
	platformUserID := c.Query("user_id")
	if platform == "" || platformUserID == "" {
		return api.JSONResponse(c, fiber.StatusBadRequest, "platform and user_id are required")
	}
	u, err := h.svc.client.User.
		Query().
		Where(user.PlatformEQ(platform), user.UserIDEQ(platformUserID)).
		First(ctx)
	if err != nil {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrUserNotFound)
	}
	return api.JSONResponse(c, fiber.StatusOK, "ok", UserResponse{
		ID:        u.ID,
		Platform:  u.Platform,
		UserID:    u.UserID,
		BanState:  u.BanState,
		BanReason: u.BanReason,
	})
}

func (h *UserHandler) GetUserByID(c fiber.Ctx) error {
	ctx := context.Background()
	harukiUserID := fiber.Params[int](c, "haruki_user_id", 0)
	if harukiUserID == 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidHarukiUserID)
	}
	u, err := h.svc.client.User.Get(ctx, harukiUserID)
	if err != nil {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrUserNotFound)
	}
	return api.JSONResponse(c, fiber.StatusOK, "ok", UserResponse{
		ID:        u.ID,
		Platform:  u.Platform,
		UserID:    u.UserID,
		BanState:  u.BanState,
		BanReason: u.BanReason,
	})
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
	ctx := context.Background()
	var req CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	if !api.ValidateStringLength(req.Platform, api.MaxPlatformLength) {
		return api.JSONResponse(c, fiber.StatusBadRequest, "platform too long")
	}
	if !api.ValidateStringLength(req.UserID, api.MaxUserIDLength) {
		return api.JSONResponse(c, fiber.StatusBadRequest, "user_id too long")
	}
	existing, _ := h.svc.client.User.
		Query().
		Where(user.PlatformEQ(req.Platform), user.UserIDEQ(req.UserID)).
		First(ctx)
	if existing != nil {
		return api.JSONResponse(c, fiber.StatusOK, "ok", UserResponse{
			ID:        existing.ID,
			Platform:  existing.Platform,
			UserID:    existing.UserID,
			BanState:  existing.BanState,
			BanReason: existing.BanReason,
		})
	}
	var newID int
	for i := 0; i < 10; i++ {
		id, err := generateUserID()
		if err != nil {
			return api.InternalError(c)
		}
		_, err = h.svc.client.User.Get(ctx, id)
		if err != nil {
			newID = id
			break
		}
	}
	if newID == 0 {
		return api.InternalError(c)
	}
	u, err := h.svc.client.User.
		Create().
		SetID(newID).
		SetPlatform(req.Platform).
		SetUserID(req.UserID).
		SetBanState(false).
		Save(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	return api.JSONResponse(c, fiber.StatusCreated, "User created", UserResponse{
		ID:        u.ID,
		Platform:  u.Platform,
		UserID:    u.UserID,
		BanState:  u.BanState,
		BanReason: u.BanReason,
	})
}

func (h *UserHandler) UpdateBan(c fiber.Ctx) error {
	ctx := context.Background()
	harukiUserID := fiber.Params[int](c, "haruki_user_id", 0)
	if harukiUserID == 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidHarukiUserID)
	}
	var req UpdateBanRequest
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	if !api.ValidateStringLength(req.BanReason, api.MaxReasonLength) {
		return api.JSONResponse(c, fiber.StatusBadRequest, "ban_reason too long")
	}
	u, err := h.svc.client.User.Get(ctx, harukiUserID)
	if err != nil {
		return api.JSONResponse(c, fiber.StatusNotFound, api.ErrUserNotFound)
	}
	updated, err := u.Update().
		SetBanState(req.BanState).
		SetBanReason(req.BanReason).
		Save(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	return api.JSONResponse(c, fiber.StatusOK, "Ban state updated", UserResponse{
		ID:        updated.ID,
		Platform:  updated.Platform,
		UserID:    updated.UserID,
		BanState:  updated.BanState,
		BanReason: updated.BanReason,
	})
}

func RegisterUsersRoutes(app *fiber.App, client *users.Client) {
	svc := NewUserService(client)
	h := NewUserHandler(svc)
	r := app.Group("/user", api.VerifyAPIAuthorization())

	r.Get("/", h.GetUser)
	r.Get("/:haruki_user_id", h.GetUserByID)
	r.Post("/", h.CreateUser)
	r.Patch("/:haruki_user_id/ban", h.UpdateBan)
}
