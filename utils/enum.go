package utils

type AliasType string

const (
	AliasTypeMusic     AliasType = "music"
	AliasTypeCharacter AliasType = "character"
)

type BindingServer string

const (
	BindingServerJP BindingServer = "jp"
	BindingServerEN BindingServer = "en"
	BindingServerTW BindingServer = "tw"
	BindingServerKR BindingServer = "kr"
	BindingServerCN BindingServer = "cn"
)

type DefaultBindingServer string

const (
	DefaultBindingServerJP      DefaultBindingServer = "jp"
	DefaultBindingServerEN      DefaultBindingServer = "en"
	DefaultBindingServerTW      DefaultBindingServer = "tw"
	DefaultBindingServerKR      DefaultBindingServer = "kr"
	DefaultBindingServerCN      DefaultBindingServer = "cn"
	DefaultBindingServerDefault DefaultBindingServer = "default"
)
