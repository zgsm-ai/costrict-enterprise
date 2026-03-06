package config

import (
	"bytes"
	"costrict-keeper/internal/env"
	"costrict-keeper/internal/logger"
	"costrict-keeper/internal/utils"
	"encoding/json"
	"html/template"
	"log"
	"os"
	"path/filepath"
)

type MidnightRooster struct {
	StartHour int `json:"start_hour,omitempty"`
	EndHour   int `json:"end_hour,omitempty"`
}

type MaintainInterval struct {
	Monitoring    int `json:"monitoring,omitempty"`
	MetricsReport int `json:"metrics_report,omitempty"`
	LogReport     int `json:"log_report,omitempty"`
}

type ServiceConfig struct {
	MinPort int `json:"min_port,omitempty"`
	MaxPort int `json:"max_port,omitempty"`
}

type TunnelConfig struct {
	ProcessName string   `json:"process_name,omitempty"`
	Command     string   `json:"command,omitempty"`
	Args        []string `json:"args,omitempty"`
	Timeout     int      `json:"timeout,omitempty"`
}

type ComponentConfig struct {
	PublicKey string `json:"public_key,omitempty"`
}

/**
 * Logging configuration
 * @property {string} level - Log level (debug/info/warn/error)
 * @property {string} path - Log file path
 * @property {int64} maxSize - Maximum log file size in bytes (default: 5242880, which is 5MB)
 * @property {int} backup - Maximum number of log backup files (default: 1)
 */
type LogConfig struct {
	Level   string `json:"level"`
	Path    string `json:"path"`
	MaxSize int64  `json:"maxSize"`
	Backup  int    `json:"backup"`
}

type CloudConfig struct {
	PushgatewayUrl string `json:"pushgateway_url,omitempty"`
	TunManagerUrl  string `json:"tunman_url,omitempty"`
	TunnelUrl      string `json:"tunnel_url,omitempty"`
	UpgradeUrl     string `json:"upgrade_url,omitempty"`
	LogUrl         string `json:"log_url,omitempty"`
}

type AppConfig struct {
	Listen    string           `json:"listen,omitempty"`
	Midnight  MidnightRooster  `json:"midnight,omitempty"`
	Interval  MaintainInterval `json:"interval,omitempty"`
	Service   ServiceConfig    `json:"service,omitempty"`
	Tunnel    TunnelConfig     `json:"tunnel,omitempty"`
	Component ComponentConfig  `json:"component,omitempty"`
	Cloud     CloudConfig      `json:"cloud,omitempty"`
	Log       LogConfig        `json:"log,omitempty"`
}

var (
	appConfig   *AppConfig
	cloudConfig *CloudConfig
)

/**
 * Load application configuration from JSON file
 * @param {string} configPath - Path to configuration file
 * @returns {error} Returns error if loading fails, nil on success
 */
func (cfg *AppConfig) loadConfig(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var newConfig AppConfig
	if err := json.NewDecoder(file).Decode(&newConfig); err != nil {
		return err
	}
	*cfg = newConfig
	return nil
}

func (cfg *AppConfig) correctConfig() {
	if cfg.Listen == "" {
		cfg.Listen = "localhost:8999"
	}
	if cfg.Midnight.StartHour == 0 {
		cfg.Midnight.StartHour = 3
	}
	if cfg.Midnight.EndHour == 0 {
		cfg.Midnight.EndHour = 5
	}
	if cfg.Interval.Monitoring == 0 {
		cfg.Interval.Monitoring = 300
	}
	if cfg.Interval.MetricsReport == 0 {
		cfg.Interval.MetricsReport = 300
	}
	if cfg.Interval.LogReport == 0 {
		cfg.Interval.LogReport = 600
	}
	// LogReportInterval 默认为 0，表示不上报日志
	if cfg.Cloud.PushgatewayUrl == "" {
		cfg.Cloud.PushgatewayUrl = "{{.BaseUrl}}/pushgateway"
	}
	if cfg.Cloud.UpgradeUrl == "" {
		cfg.Cloud.UpgradeUrl = "{{.BaseUrl}}/costrict"
	}
	if cfg.Cloud.TunnelUrl == "" {
		cfg.Cloud.TunnelUrl = "{{.BaseUrl}}/ws"
	}
	if cfg.Cloud.TunManagerUrl == "" {
		cfg.Cloud.TunManagerUrl = "{{.BaseUrl}}/tunnel-manager/api/v1"
	}
	if cfg.Cloud.LogUrl == "" {
		cfg.Cloud.LogUrl = "{{.BaseUrl}}/client-manager/api/v1/logs"
	}
	if cfg.Service.MinPort == 0 {
		cfg.Service.MinPort = 9000
	}
	if cfg.Service.MaxPort == 0 {
		cfg.Service.MaxPort = cfg.Service.MinPort + 1000
	}
	if cfg.Tunnel.ProcessName == "" {
		cfg.Tunnel.ProcessName = "cotun"
	}
	if cfg.Tunnel.Command == "" {
		cfg.Tunnel.Command = "{{.ProcessPath}}"
	}
	if len(cfg.Tunnel.Args) == 0 {
		cfg.Tunnel.Args = []string{
			"--auth",
			"costrict:zgsm@costrict.ai",
			"--tls-skip-verify",
			"--server",
			"{{.RemoteAddr}}",
			"--client-port",
			"{{.LocalPort}}",
			"--mapping-port",
			"{{.MappingPort}}",
		}
	}
	// 设置默认日志配置
	if cfg.Log.Level == "" {
		cfg.Log.Level = "debug"
	}
	if cfg.Log.Path == "" {
		cfg.Log.Path = "console" // 默认输出到控制台
	}
	if cfg.Log.MaxSize == 0 {
		cfg.Log.MaxSize = 1 * 1024 * 1024 // 默认1M
	}
	if cfg.Log.Backup == 0 {
		cfg.Log.Backup = 1
	}
}

func expandUrl(baseUrl string, pattern string) (string, error) {
	tpl, err := template.New("url").Parse(pattern)
	if err != nil {
		return "", err
	}
	var data struct{ BaseUrl string }
	data.BaseUrl = baseUrl
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func expandCloudConfig(cloud *CloudConfig) *CloudConfig {
	expand := CloudConfig{}
	baseUrl := GetBaseURL()
	if baseUrl == "" {
		baseUrl = "https://zgsm.sangfor.com"
	}
	var err error
	expand.PushgatewayUrl, err = expandUrl(baseUrl, cloud.PushgatewayUrl)
	if err != nil {
		logger.Errorf("Invalid pushgateway_url: %s", cloud.PushgatewayUrl)
		return nil
	}
	expand.TunManagerUrl, err = expandUrl(baseUrl, cloud.TunManagerUrl)
	if err != nil {
		logger.Errorf("Invalid tunmanager_url: %s", cloud.TunManagerUrl)
		return nil
	}
	expand.TunnelUrl, err = expandUrl(baseUrl, cloud.TunnelUrl)
	if err != nil {
		logger.Errorf("Invalid tunnel_url: %s", cloud.TunnelUrl)
		return nil
	}
	expand.UpgradeUrl, err = expandUrl(baseUrl, cloud.UpgradeUrl)
	if err != nil {
		logger.Errorf("Invalid upgrade_url: %s", cloud.UpgradeUrl)
		return nil
	}
	expand.LogUrl, err = expandUrl(baseUrl, cloud.LogUrl)
	if err != nil {
		logger.Errorf("Invalid log_url: %s", cloud.LogUrl)
		return nil
	}
	return &expand
}

func LoadConfig(ignoreError bool) error {
	var cfg AppConfig
	configPath := filepath.Join(env.CostrictDir, "config", "costrict.json")
	if err := cfg.loadConfig(configPath); err != nil {
		if !ignoreError {
			return err
		}
	}
	cfg.correctConfig()
	utils.SetAvailablePortRange(cfg.Service.MinPort, cfg.Service.MaxPort)
	cloudConfig = expandCloudConfig(&cfg.Cloud)
	appConfig = &cfg
	return nil
}

/**
 * Load configuration from specified path
 * @returns {error} Returns error if loading fails, nil on success
 */
func ReloadConfig(ignoreError bool) error {
	if err := fetchRemoteConfig("costrict-config"); err != nil {
		if !ignoreError {
			return err
		}
	}
	return LoadConfig(ignoreError)
}

/**
 * App configuration instance
 * @returns {AppConfig} Returns configuration instance
 */
func App() *AppConfig {
	if appConfig == nil {
		log.Fatal("Must run config.LoadConfig first")
		return nil
	}
	return appConfig
}

func Cloud() *CloudConfig {
	if cloudConfig == nil {
		log.Fatal("Must run config.LoadConfig first")
		return nil
	}
	return cloudConfig
}
