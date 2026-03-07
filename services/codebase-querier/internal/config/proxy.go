package config

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// ProxyConfig 代理配置
type ProxyConfig struct {
	Mode           string            `json:"mode" yaml:"mode"`     // 代理模式: rewrite, full_path
	Routes         []RouteConfig     `json:"routes" yaml:"routes"` // 路由规则数组
	Rewrite        RewriteConfig     `json:"rewrite" yaml:"rewrite"`
	Headers        HeadersConfig     `json:"headers" yaml:"headers"`
	PortManagerURL string            `json:"port_manager_url" yaml:"port_manager_url"` // 端口管理器URL
	DynamicPort    bool              `json:"dynamic_port" yaml:"dynamic_port"`         // 是否启用动态端口
	PortManager    PortManagerConfig `json:"port_manager" yaml:"port_manager"`         // 端口管理器配置
	ForwardURL     string            `json:"forward_url" yaml:"forward_url"`           // 转发地址
	// 基于请求头的转发配置
	HeaderBasedForward HeaderBasedForwardConfig `json:"header_based_forward" yaml:"header_based_forward"` // 基于请求头的转发配置
}

// HeaderBasedForwardConfig 基于请求头的转发配置
type HeaderBasedForwardConfig struct {
	Enabled    bool                           `json:"enabled" yaml:"enabled"`         // 是否启用基于请求头的转发
	HeaderName string                         `json:"header_name" yaml:"header_name"` // 请求头名称
	Paths      []HeaderBasedForwardPathConfig `json:"paths" yaml:"paths"`             // 多路径配置数组
}

// HeaderBasedForwardPathConfig 基于请求头的转发路径配置
type HeaderBasedForwardPathConfig struct {
	Path             string `json:"path" yaml:"path"`                             // 目标路径
	WithHeaderURL    string `json:"with_header_url" yaml:"with_header_url"`       // 有请求头时的转发地址
	WithoutHeaderURL string `json:"without_header_url" yaml:"without_header_url"` // 无请求头时的转发地址
}

// PortManagerConfig 端口管理器配置
type PortManagerConfig struct {
	URL                 string        `json:"url" yaml:"url"`                                         // 端口管理器URL
	ForwardURL          string        `json:"forward_url" yaml:"forward_url"`                         // 转发地址
	Timeout             time.Duration `json:"timeout" yaml:"timeout"`                                 // 请求超时时间
	CacheExp            time.Duration `json:"cache_exp" yaml:"cache_exp"`                             // 缓存过期时间
	MaxIdleConns        int           `json:"max_idle_conns" yaml:"max_idle_conns"`                   // 最大空闲连接数
	MaxIdleConnsPerHost int           `json:"max_idle_conns_per_host" yaml:"max_idle_conns_per_host"` // 每个主机的最大空闲连接数
	IdleConnTimeout     time.Duration `json:"idle_conn_timeout" yaml:"idle_conn_timeout"`             // 空闲连接超时时间
}

// RouteConfig 路由配置
type RouteConfig struct {
	PathPrefix string       `json:"path_prefix" yaml:"path_prefix"` // 路径前缀
	Target     TargetConfig `json:"target" yaml:"target"`           // 目标服务配置
}

// TargetConfig 目标服务配置
type TargetConfig struct {
	URL     string        `json:"url" yaml:"url"`
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
}

// UnmarshalYAML 自定义YAML解析方法，支持直接解析时间字符串（如"30s"）
func (t *TargetConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux struct {
		URL     string `json:"url" yaml:"url"`
		Timeout string `json:"timeout" yaml:"timeout"`
	}
	if err := unmarshal(&aux); err != nil {
		return err
	}

	t.URL = aux.URL

	// 如果timeout是字符串格式（如"30s"），则解析为time.Duration
	if aux.Timeout != "" {
		d, err := time.ParseDuration(aux.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout format: %v", err)
		}
		t.Timeout = d
	} else {
		// 默认超时时间
		t.Timeout = 30 * time.Second
	}

	return nil
}

// RewriteConfig 路径重写配置
type RewriteConfig struct {
	Enabled bool          `json:"enabled" yaml:"enabled"`
	Rules   []RewriteRule `json:"rules" yaml:"rules"`
}

// RewriteRule 重写规则
type RewriteRule struct {
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
}

// HeadersConfig Header配置
type HeadersConfig struct {
	PassThrough bool              `json:"pass_through" yaml:"pass_through"`
	Exclude     []string          `json:"exclude" yaml:"exclude"`
	Override    map[string]string `json:"override" yaml:"override"`
}

// 代理模式常量
const (
	ProxyModeRewrite  = "rewrite"   // 路径重写模式
	ProxyModeFullPath = "full_path" // 全路径转发模式
)

// Validate 验证代理配置
func (c *ProxyConfig) Validate() error {
	if c.Mode == "" {
		c.Mode = ProxyModeRewrite // 默认为rewrite模式，保持向后兼容
	}

	if c.Mode != ProxyModeRewrite && c.Mode != ProxyModeFullPath {
		return fmt.Errorf("invalid proxy mode: %s, must be %s or %s", c.Mode, ProxyModeRewrite, ProxyModeFullPath)
	}

	// 如果启用动态端口，验证端口管理器URL
	if c.DynamicPort {
		// 检查新的端口管理器配置
		if c.PortManager.URL == "" {
			// 如果新配置为空，则使用旧的PortManagerURL作为后备
			if c.PortManagerURL == "" {
				return errors.New("port_manager.url is required when dynamic_port is enabled")
			}
			c.PortManager.URL = c.PortManagerURL
		}
		if _, err := url.Parse(c.PortManager.URL); err != nil {
			return fmt.Errorf("invalid port_manager.url: %w", err)
		}
		if c.PortManager.Timeout <= 0 {
			c.PortManager.Timeout = 10 * time.Second
		}
		if c.PortManager.CacheExp <= 0 {
			c.PortManager.CacheExp = 5 * time.Minute
		}
		if c.PortManager.MaxIdleConns <= 0 {
			c.PortManager.MaxIdleConns = 10
		}
		if c.PortManager.MaxIdleConnsPerHost <= 0 {
			c.PortManager.MaxIdleConnsPerHost = 5
		}
		if c.PortManager.IdleConnTimeout <= 0 {
			c.PortManager.IdleConnTimeout = 30 * time.Second
		}
	}

	// 验证路由配置
	if len(c.Routes) == 0 && !c.DynamicPort {
		return errors.New("at least one route is required when dynamic_port is disabled")
	}

	for i, route := range c.Routes {
		if route.PathPrefix == "" {
			return fmt.Errorf("route[%d] path_prefix is required", i)
		}
		if route.Target.URL == "" {
			return fmt.Errorf("route[%d] target URL is required", i)
		}
		if _, err := url.Parse(route.Target.URL); err != nil {
			return fmt.Errorf("route[%d] invalid target URL: %w", i, err)
		}
		if route.Target.Timeout <= 0 {
			c.Routes[i].Target.Timeout = 30 * time.Second
		}
	}

	// full_path模式下禁用rewrite
	if c.Mode == ProxyModeFullPath {
		c.Rewrite.Enabled = false
	}

	for _, rule := range c.Rewrite.Rules {
		if rule.From == "" {
			return errors.New("rewrite rule 'from' cannot be empty")
		}
	}

	// 验证基于请求头的转发配置
	if c.HeaderBasedForward.Enabled {
		if c.HeaderBasedForward.HeaderName == "" {
			return errors.New("header_based_forward.header_name is required when header_based_forward.enabled is true")
		}
		if len(c.HeaderBasedForward.Paths) == 0 {
			return errors.New("header_based_forward.paths is required when header_based_forward.enabled is true")
		}
		for i, pathConfig := range c.HeaderBasedForward.Paths {
			if pathConfig.Path == "" {
				return fmt.Errorf("header_based_forward.paths[%d].path is required", i)
			}
			if pathConfig.WithHeaderURL == "" {
				return fmt.Errorf("header_based_forward.paths[%d].with_header_url is required", i)
			}
			if pathConfig.WithoutHeaderURL == "" {
				return fmt.Errorf("header_based_forward.paths[%d].without_header_url is required", i)
			}
		}
	}

	return nil
}

// DefaultProxyConfig 返回默认配置
func DefaultProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		Mode:        ProxyModeRewrite, // 默认为rewrite模式，保持向后兼容
		DynamicPort: false,            // 默认不启用动态端口
		Routes: []RouteConfig{
			{
				PathPrefix: "/",
				Target: TargetConfig{
					Timeout: 30 * time.Second,
				},
			},
		},
		Rewrite: RewriteConfig{
			Enabled: false,
		},
		Headers: HeadersConfig{
			PassThrough: true,
			Exclude:     []string{},
			Override:    make(map[string]string),
		},
		PortManager: PortManagerConfig{
			ForwardURL:          "http://10.233.23.31", // 默认转发地址
			Timeout:             10 * time.Second,
			CacheExp:            5 * time.Minute,
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
		HeaderBasedForward: HeaderBasedForwardConfig{
			Enabled:    false,                // 默认不启用基于请求头的转发
			HeaderName: "X-Costrict-Version", // 默认请求头名称
			Paths: []HeaderBasedForwardPathConfig{
				{
					Path:             "/codebase-embedder/api/v1/search/semantic", // 默认目标路径
					WithHeaderURL:    "codebase-embedder/api/v1/search/semantic",  // 有请求头时的转发地址
					WithoutHeaderURL: "/codebase-index/api/v1/search/semantic",    // 无请求头时的转发地址
				},
			},
		},
	}
}
