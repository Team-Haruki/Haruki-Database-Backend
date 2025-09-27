package utils

import "fmt"

type AliasType string

const (
	AliasTypeMusic     AliasType = "music"
	AliasTypeCharacter AliasType = "character"
)

func ParseAliasType(t string) (AliasType, error) {
	switch AliasType(t) {
	case AliasTypeMusic,
		AliasTypeCharacter:
		return AliasType(t), nil
	default:
		return "", fmt.Errorf("invalid alias type: %s", s)
	}
}

type BindingServer string

const (
	BindingServerJP BindingServer = "jp"
	BindingServerEN BindingServer = "en"
	BindingServerTW BindingServer = "tw"
	BindingServerKR BindingServer = "kr"
	BindingServerCN BindingServer = "cn"
)

func ParseBindingServer(s string) (BindingServer, error) {
	switch BindingServer(s) {
	case BindingServerJP,
		BindingServerEN,
		BindingServerTW,
		BindingServerKR,
		BindingServerCN:
		return BindingServer(s), nil
	default:
		return "", fmt.Errorf("invalid server: %s", s)
	}
}

type DefaultBindingServer string

const (
	DefaultBindingServerJP      DefaultBindingServer = "jp"
	DefaultBindingServerEN      DefaultBindingServer = "en"
	DefaultBindingServerTW      DefaultBindingServer = "tw"
	DefaultBindingServerKR      DefaultBindingServer = "kr"
	DefaultBindingServerCN      DefaultBindingServer = "cn"
	DefaultBindingServerDefault DefaultBindingServer = "default"
)

func ParseDefaultBindingServer(s string) (DefaultBindingServer, error) {
	switch DefaultBindingServer(s) {
	case DefaultBindingServerJP,
		DefaultBindingServerEN,
		DefaultBindingServerTW,
		DefaultBindingServerKR,
		DefaultBindingServerCN,
		DefaultBindingServerDefault:
		return DefaultBindingServer(s), nil
	default:
		return "", fmt.Errorf("invalid server: %s", s)
	}
}
