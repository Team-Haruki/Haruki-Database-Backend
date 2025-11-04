package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type BackendConfig struct {
	Host                string        `yaml:"host"`
	Port                int           `yaml:"port"`
	SSL                 bool          `yaml:"ssl"`
	SSLCert             string        `yaml:"ssl_cert"`
	SSLKey              string        `yaml:"ssl_key"`
	LogLevel            string        `yaml:"log_level"`
	MainLogFile         string        `yaml:"main_log_file"`
	AccessLog           string        `yaml:"access_log"`
	APICacheTTL         time.Duration `yaml:"api_cache_ttl"`
	AccessLogPath       string        `yaml:"access_log_path"`
	AcceptAuthorization string        `yaml:"accept_authorization"`
	AcceptUserAgent     string        `yaml:"accept_user_agent"`
	EnableTrustProxy    bool          `yaml:"enable_trust_proxy"`
	TrustProxies        []string      `yaml:"trust_proxies"`
	ProxyHeader         string        `yaml:"proxy_header"`
}

type ChunithmConfig struct {
	Enabled       bool   `yaml:"enabled"`
	MusicDBType   string `yaml:"music_db_type"`
	MusicDBURL    string `yaml:"music_db_url"`
	BindingDBType string `yaml:"binding_db_type"`
	BindingDBURL  string `yaml:"binding_db_url"`
}

type PJSKConfig struct {
	Enabled bool   `yaml:"enabled"`
	DBType  string `yaml:"db_type"`
	DBURL   string `yaml:"db_url"`
}

type CensorConfig struct {
	BaiduAPIKey  string `yaml:"baidu_api_key"`
	BaiduSecret  string `yaml:"baidu_secret"`
	CensorDBType string `yaml:"censor_db_type"`
	CensorDBURL  string `yaml:"censor_db_url"`
}

type HarukiBotDBConfig struct {
	DBType              string `yaml:"db_type"`
	DBURL               string `yaml:"db_url"`
	RegisterVerifyToken string `yaml:"register_verify_token"`
	CredentialSignToken string `yaml:"credential_sign_token"`
	LoginSignToken      string `yaml:"login_sign_token"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type Config struct {
	Backend     BackendConfig     `yaml:"backend"`
	Chunithm    ChunithmConfig    `yaml:"chunithm"`
	PJSK        PJSKConfig        `yaml:"pjsk"`
	Censor      CensorConfig      `yaml:"censor"`
	HarukiBotDB HarukiBotDBConfig `yaml:"haruki_bot"`
	Redis       RedisConfig       `yaml:"redis"`
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
