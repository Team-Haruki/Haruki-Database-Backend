package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	AcceptAuthorization string `yaml:"accept_authorization"`
	AcceptUserAgent     string `yaml:"accept_user_agent"`
}

type ChunithmConfig struct {
	Enabled      bool   `yaml:"enabled"`
	MusicDBURL   string `yaml:"music_db_url"`
	BindingDBURL string `yaml:"binding_db_url"`
}

type PJSKConfig struct {
	Enabled bool   `yaml:"enabled"`
	DBURL   string `yaml:"db_url"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type Config struct {
	App      AppConfig      `yaml:"app"`
	Chunithm ChunithmConfig `yaml:"chunithm"`
	PJSK     PJSKConfig     `yaml:"pjsk"`
	Redis    RedisConfig    `yaml:"redis"`
}

var Cfg Config

func LoadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(data, &Cfg)
	if err != nil {
		log.Fatalf("failed to unmarshal config file: %v", err)
	}
}
