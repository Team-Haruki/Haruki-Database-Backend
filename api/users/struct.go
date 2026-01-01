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

type UpdateFeatureBanRequest struct {
	BanState  bool   `json:"ban_state"`
	BanReason string `json:"ban_reason"`
}

type UserResponse struct {
	ID                     int    `json:"id"`
	Platform               string `json:"platform"`
	UserID                 string `json:"user_id"`
	BanState               bool   `json:"ban_state"`
	BanReason              string `json:"ban_reason,omitempty"`
	PjskBanState           bool   `json:"pjsk_ban_state"`
	PjskBanReason          string `json:"pjsk_ban_reason,omitempty"`
	ChunithmBanState       bool   `json:"chunithm_ban_state"`
	ChunithmBanReason      string `json:"chunithm_ban_reason,omitempty"`
	PjskMainBanState       bool   `json:"pjsk_main_ban_state"`
	PjskMainBanReason      string `json:"pjsk_main_ban_reason,omitempty"`
	PjskRankingBanState    bool   `json:"pjsk_ranking_ban_state"`
	PjskRankingBanReason   string `json:"pjsk_ranking_ban_reason,omitempty"`
	PjskAliasBanState      bool   `json:"pjsk_alias_ban_state"`
	PjskAliasBanReason     string `json:"pjsk_alias_ban_reason,omitempty"`
	PjskMysekaiBanState    bool   `json:"pjsk_mysekai_ban_state"`
	PjskMysekaiBanReason   string `json:"pjsk_mysekai_ban_reason,omitempty"`
	ChunithmMainBanState   bool   `json:"chunithm_main_ban_state"`
	ChunithmMainBanReason  string `json:"chunithm_main_ban_reason,omitempty"`
	ChunithmAliasBanState  bool   `json:"chunithm_alias_ban_state"`
	ChunithmAliasBanReason string `json:"chunithm_alias_ban_reason,omitempty"`
}

type UserService struct {
	client *users.Client
}

type UserHandler struct {
	svc *UserService
}
