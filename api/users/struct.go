package users

import (
	"haruki-database/database/schema/users"
)

type CreateUserRequest struct {
	Platform string `json:"platform"`
	UserID   string `json:"user_id"`
}

type UpdateBanRequest struct {
	BanState  bool   `json:"ban_state"`
	BanReason string `json:"ban_reason"`
}

type UserResponse struct {
	ID        int    `json:"id"`
	Platform  string `json:"platform"`
	UserID    string `json:"user_id"`
	BanState  bool   `json:"ban_state"`
	BanReason string `json:"ban_reason,omitempty"`
}

type UserService struct {
	client *users.Client
}

type UserHandler struct {
	svc *UserService
}
