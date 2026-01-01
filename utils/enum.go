package utils

import "fmt"

// ================= Validation Constants =================

const (
	MaxAliasLength    = 100
	MaxUserIDLength   = 50
	MaxServerLength   = 20
	MaxReasonLength   = 255
	MaxOptionLength   = 50
	MaxValueLength    = 50
	MaxPlatformLength = 20
)

// ================= Error Messages =================

const (
	ErrInvalidRequest      = "Invalid request"
	ErrInvalidUserID       = "Invalid user_id"
	ErrInvalidHarukiUserID = "Invalid haruki_user_id"
	ErrUserNotFound        = "User not found"
	ErrAliasNotFound       = "Alias not found"
	ErrBindingNotFound     = "Binding not found"
	ErrPreferenceNotFound  = "Preference not found"
	ErrPermissionDenied    = "Permission denied"
	ErrAlreadyExists       = "Already exists"
	ErrInternalServer      = "Internal server error"
)

// ================= Alias Type Enum =================

type AliasType string

const (
	AliasTypeMusic     AliasType = "music"
	AliasTypeCharacter AliasType = "character"
)

// Valid returns true if the alias type is valid
func (a AliasType) Valid() bool {
	switch a {
	case AliasTypeMusic, AliasTypeCharacter:
		return true
	default:
		return false
	}
}

func ParseAliasType(t string) (AliasType, error) {
	at := AliasType(t)
	if !at.Valid() {
		return "", fmt.Errorf("invalid alias type: %s", t)
	}
	return at, nil
}

// ================= Binding Server Enum =================

type BindingServer string

const (
	BindingServerJP BindingServer = "jp"
	BindingServerEN BindingServer = "en"
	BindingServerTW BindingServer = "tw"
	BindingServerKR BindingServer = "kr"
	BindingServerCN BindingServer = "cn"
)

// Valid returns true if the binding server is valid
func (s BindingServer) Valid() bool {
	switch s {
	case BindingServerJP, BindingServerEN, BindingServerTW, BindingServerKR, BindingServerCN:
		return true
	default:
		return false
	}
}

func ParseBindingServer(s string) (BindingServer, error) {
	bs := BindingServer(s)
	if !bs.Valid() {
		return "", fmt.Errorf("invalid server: %s", s)
	}
	return bs, nil
}

// ================= Default Binding Server Enum =================

type DefaultBindingServer string

const (
	DefaultBindingServerJP      DefaultBindingServer = "jp"
	DefaultBindingServerEN      DefaultBindingServer = "en"
	DefaultBindingServerTW      DefaultBindingServer = "tw"
	DefaultBindingServerKR      DefaultBindingServer = "kr"
	DefaultBindingServerCN      DefaultBindingServer = "cn"
	DefaultBindingServerDefault DefaultBindingServer = "default"
)

// Valid returns true if the default binding server is valid
func (s DefaultBindingServer) Valid() bool {
	switch s {
	case DefaultBindingServerJP, DefaultBindingServerEN, DefaultBindingServerTW,
		DefaultBindingServerKR, DefaultBindingServerCN, DefaultBindingServerDefault:
		return true
	default:
		return false
	}
}

func ParseDefaultBindingServer(s string) (DefaultBindingServer, error) {
	dbs := DefaultBindingServer(s)
	if !dbs.Valid() {
		return "", fmt.Errorf("invalid server: %s", s)
	}
	return dbs, nil
}
