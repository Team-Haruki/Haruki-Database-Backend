package api

import (
	"haruki-database/utils"
	"haruki-database/utils/types"
)

type AliasToIDResponse = types.AliasToIDResponse
type AliasListResponse = types.AliasListResponse
type AliasRequest = types.AliasRequest
type RejectRequest = types.RejectRequest
type PendingAlias = types.PendingAlias
type RejectedAlias = types.RejectedAlias

const UserContextKey = "haruki_user"

const (
	MaxAliasLength    = utils.MaxAliasLength
	MaxUserIDLength   = utils.MaxUserIDLength
	MaxServerLength   = utils.MaxServerLength
	MaxReasonLength   = utils.MaxReasonLength
	MaxOptionLength   = utils.MaxOptionLength
	MaxValueLength    = utils.MaxValueLength
	MaxPlatformLength = utils.MaxPlatformLength
)

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
)
