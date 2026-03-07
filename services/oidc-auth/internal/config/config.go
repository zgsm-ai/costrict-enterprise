package config

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/zgsm-ai/oidc-auth/pkg/log"
)

type AppConfig struct {
	Server       Server                    `json:"Server" mapstructure:"Server" validate:"required"`
	Log          LogConfig                 `json:"log" mapstructure:"log" validate:"required"`
	Database     DatabaseConfig            `json:"database" mapstructure:"database" validate:"required"`
	GithubConfig GithubStarConfig          `json:"syncStar" mapstructure:"syncStar" validate:"required"`
	Encrypt      EncryptConfig             `json:"encrypt" mapstructure:"encrypt" validate:"required"`
	SMS          SMSConfig                 `json:"sms" mapstructure:"sms" validate:"required"`
	Providers    map[string]ProviderConfig `json:"providers" mapstructure:"providers"`
	QuotaManager QuotaConfig               `json:"quotaManager" mapstructure:"quotaManager"`
	Redirect     RedirectConfig            `json:"redirect" mapstructure:"redirect"`
}

type Server struct {
	ServerPort string            `json:"serverPort" mapstructure:"serverPort" validate:"required,numeric"`
	BaseURL    string            `json:"baseURL" mapstructure:"baseURL"`
	HTTP       *HTTPClientConfig `json:"http" mapstructure:"http" validate:"required"`
	IsPrivate  bool              `json:"isPrivate" mapstructure:"isPrivate"`
}

type RedirectConfig struct {
	Uris map[string]string `json:"uris" mapstructure:"uris"`
}

type HTTPClientConfig struct {
	Timeout               time.Duration `json:"timeout" mapstructure:"timeout" validate:"gte=0"`
	DialTimeout           time.Duration `json:"dialTimeout" mapstructure:"dialTimeout" validate:"gte=0"`
	KeepAlive             time.Duration `json:"keepAlive" mapstructure:"keepAlive" validate:"gte=0"`
	TLSHandshakeTimeout   time.Duration `json:"tlsHandshakeTimeout" mapstructure:"tlsHandshakeTimeout" validate:"gte=0"`
	ResponseHeaderTimeout time.Duration `json:"responseHeaderTimeout" mapstructure:"responseHeaderTimeout" validate:"gte=0"`
	MaxIdleConns          int           `json:"maxIdleConns" mapstructure:"maxIdleConns" validate:"gte=0"`
	MaxIdleConnsPerHost   int           `json:"maxIdleConnsPerHost" mapstructure:"maxIdleConnsPerHost" validate:"gte=0"`
	IdleConnTimeout       time.Duration `json:"idleConnTimeout" mapstructure:"idleConnTimeout" validate:"gte=0"`
}

type LogConfig struct {
	Level      string `json:"level" mapstructure:"level" validate:"required,oneof=debug info warn error fatal"`
	Filename   string `json:"filename" mapstructure:"filename"`
	MaxSize    int    `json:"maxSize" mapstructure:"maxSize"`
	MaxBackups int    `json:"maxBackups" mapstructure:"maxBackups"`
	MaxAge     int    `json:"maxAge" mapstructure:"maxAge"`
	Compress   bool   `json:"compress" mapstructure:"compress"`
}

type DatabaseConfig struct {
	Type         string `json:"type" mapstructure:"type" validate:"required,oneof=mysql postgres"`
	Host         string `json:"host" mapstructure:"host" validate:"required"`
	Port         int    `json:"port" mapstructure:"port" validate:"required,min=1,max=65535"`
	Username     string `json:"username" mapstructure:"username" validate:"required"`
	Password     string `json:"password" mapstructure:"password" validate:"required"`
	DBName       string `json:"dbname" mapstructure:"dbname" validate:"required"`
	MaxIdleConns int    `json:"maxIdleConns" mapstructure:"maxIdleConns" validate:"omitempty,min=1"`
	MaxOpenConns int    `json:"maxOpenConns" mapstructure:"maxOpenConns" validate:"omitempty,min=1"`
}

type GithubStarConfig struct {
	Enabled       bool          `json:"enabled" mapstructure:"enabled" validate:"required"`
	PersonalToken string        `json:"personalToken" mapstructure:"personalToken" validate:"required"`
	Owner         string        `json:"owner" mapstructure:"owner" validate:"required"`
	Repo          string        `json:"repo" mapstructure:"repo" validate:"required"`
	Interval      time.Duration `json:"interval" mapstructure:"interval" validate:"required"`
	HTTPClient    *http.Client
}

type EncryptConfig struct {
	PrivateKeyPath string `json:"privateKey" mapstructure:"privateKey"`
	PublicKeyPath  string `json:"publicKey" mapstructure:"publicKey"`
	AesKey         string `json:"aesKey" mapstructure:"aesKey" validate:"required"`
	EnableRsa      bool   `json:"enableRsa" mapstructure:"enableRsa"`
}

type SMSConfig struct {
	EnabledTest bool   `json:"enabledTest" mapstructure:"enabledTest" validate:"required"`
	AppID       string `json:"appID" mapstructure:"appID"`
	MchID       string `json:"mchID" mapstructure:"mchID"`
	TemplateID  string `json:"templateID" mapstructure:"templateID"`
	ApiKey      string `json:"apiKey" mapstructure:"apiKey"`
	SendURL     string `json:"sendURL" mapstructure:"sendURL"`
	HTTPClient  *http.Client
}

type ProviderConfig struct {
	ClientID     string `json:"clientID" mapstructure:"clientID"`
	ClientSecret string `json:"clientSecret" mapstructure:"clientSecret"`
	EncryptKey   string `json:"encryptKey" mapstructure:"encryptKey"`
	BaseURL      string `json:"baseURL" mapstructure:"baseURL"`
	InternalURL  string `json:"internalURL" mapstructure:"internalURL"`
}

type QuotaConfig struct {
	BaseURL    string `json:"baseURL" mapstructure:"baseURL"`
	HTTPClient *http.Client
}

const (
	EnvPrefix         = ""
	DefaultConfigDir  = "."
	DefaultConfigName = "config"
	DefaultConfigType = "yaml"
)

func InitConfig(cfgFile string) (*AppConfig, error) {
	cfg := new(AppConfig)
	viper.SetDefault("serverPort", "8080")
	viper.SetDefault("log.level", "info")

	viper.SetDefault("database.maxIdleConns", 50)
	viper.SetDefault("database.maxOpenConns", 300)

	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // eg: DATABASE_HOST for database.host

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(DefaultConfigDir)
		viper.SetConfigName(DefaultConfigName)
		viper.SetConfigType(DefaultConfigType)
	}

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Info(nil, "Config file not found, using defaults and environment variables.")
		}
	} else {
		log.Info(nil, "Configuration loaded from: %s", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	log.Info(nil, "Configuration initialized and validated successfully.")
	return cfg, nil
}
