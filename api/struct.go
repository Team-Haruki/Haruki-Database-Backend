package api

import "haruki-database/utils"

// ================= Response Structs =================

type HarukiAPIResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type HarukiAPIDataResponse[T any] struct {
	HarukiAPIResponse
	Data T `json:"data,omitempty"`
}

// ================= User Info =================

type UserInfo struct {
	HarukiUserID int    `json:"haruki_user_id"`
	Platform     string `json:"platform"`
	UserID       string `json:"user_id"`
	BanState     bool   `json:"ban_state"`
	BanReason    string `json:"ban_reason,omitempty"`
}

// ================= Context Keys =================

const UserContextKey = "haruki_user"

// ================= Length Constants =================

const (
	MaxAliasLength    = utils.MaxAliasLength
	MaxUserIDLength   = utils.MaxUserIDLength
	MaxServerLength   = utils.MaxServerLength
	MaxReasonLength   = utils.MaxReasonLength
	MaxOptionLength   = utils.MaxOptionLength
	MaxValueLength    = utils.MaxValueLength
	MaxPlatformLength = utils.MaxPlatformLength
)

// ================= Error Messages =================

const (
	ErrInvalidRequest      = utils.ErrInvalidRequest
	ErrInvalidUserID       = utils.ErrInvalidUserID
	ErrInvalidHarukiUserID = utils.ErrInvalidHarukiUserID
	ErrUserNotFound        = utils.ErrUserNotFound
	ErrAliasNotFound       = utils.ErrAliasNotFound
	ErrBindingNotFound     = utils.ErrBindingNotFound
	ErrPreferenceNotFound  = utils.ErrPreferenceNotFound
	ErrPermissionDenied    = utils.ErrPermissionDenied
	ErrAlreadyExists       = utils.ErrAlreadyExists
	ErrInternalServer      = utils.ErrInternalServer
	ErrUserBanned          = "user is banned"
	ErrMissingPlatformInfo = "platform and platform_user_id are required"
)

// ================= Cache Keys =================

const (
	UserCacheKeyPrefix = "user:info:"
	UserCacheTTL       = 5 * 60
)
