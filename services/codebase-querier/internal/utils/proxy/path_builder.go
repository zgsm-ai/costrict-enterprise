package proxy

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// PathBuilder 路径构建器接口
type PathBuilder interface {
	BuildPath(originalPath string) (string, error)
}

// RewritePathBuilder 路径重写构建器
type RewritePathBuilder struct {
	rules []RewriteRule
}

// NewRewritePathBuilder 创建路径重写构建器
func NewRewritePathBuilder(rules []RewriteRule) *RewritePathBuilder {
	return &RewritePathBuilder{
		rules: rules,
	}
}

// BuildPath 根据重写规则构建路径
func (b *RewritePathBuilder) BuildPath(originalPath string) (string, error) {
	if len(b.rules) == 0 {
		return originalPath, nil
	}

	for _, rule := range b.rules {
		if strings.HasPrefix(originalPath, rule.From) {
			newPath := strings.Replace(originalPath, rule.From, rule.To, 1)
			logx.Infof("Path rewritten: %s -> %s", originalPath, newPath)
			return newPath, nil
		}
	}

	return originalPath, nil
}

// FullPathBuilder 全路径构建器
type FullPathBuilder struct {
	targetURL string
}

// NewFullPathBuilder 创建全路径构建器
func NewFullPathBuilder(targetURL string) *FullPathBuilder {
	return &FullPathBuilder{
		targetURL: targetURL,
	}
}

// BuildPath 保持原始路径不变
func (b *FullPathBuilder) BuildPath(originalPath string) (string, error) {
	if originalPath == "" {
		originalPath = "/"
	}

	// 确保路径格式正确
	if !strings.HasPrefix(originalPath, "/") {
		originalPath = "/" + originalPath
	}

	// 解析目标URL
	target, err := url.Parse(b.targetURL)
	if err != nil {
		return "", fmt.Errorf("invalid target URL: %w", err)
	}

	// 构建完整URL
	fullURL := *target
	fullURL.Path = path.Join(target.Path, originalPath)
	fullURL.RawQuery = "" // 查询参数将在转发时处理

	return fullURL.String(), nil
}

// GetPathBuilder 根据模式获取路径构建器
func GetPathBuilder(mode string, targetURL string, rules []RewriteRule) (PathBuilder, error) {
	switch mode {
	case "rewrite":
		return NewRewritePathBuilder(rules), nil
	case "full_path":
		return NewFullPathBuilder(targetURL), nil
	default:
		return nil, errors.New("invalid proxy mode")
	}
}
