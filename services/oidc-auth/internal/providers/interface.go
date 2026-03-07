package providers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/zgsm-ai/oidc-auth/internal/repository"
)

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresIn    int64     `json:"expires_in,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

type OAuthProvider interface {
	GetName() string

	GetEndpoint(isInternal bool) string

	GetAuthURL(state, redirectURL string) string

	ExchangeToken(ctx context.Context, code string) (*TokenResponse, error)

	GetUserInfo(ctx context.Context, accessToken string) (*repository.AuthUser, error)

	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)

	Update(ctx context.Context, data *repository.AuthUser) error
}

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
	InternalURL  string
	Client       *http.Client
}

type ProviderFactory interface {
	CreateProvider(config *ProviderConfig) OAuthProvider
}

type OAuthManager struct {
	factories map[string]ProviderFactory
	configs   map[string]*ProviderConfig
}

func NewOAuthManager() *OAuthManager {
	return &OAuthManager{
		factories: make(map[string]ProviderFactory),
		configs:   make(map[string]*ProviderConfig),
	}
}

func (m *OAuthManager) RegisterFactory(name string, factory ProviderFactory) {
	m.factories[name] = factory
}

func (m *OAuthManager) SetConfig(name string, config *ProviderConfig) {
	m.configs[name] = config
}

func (m *OAuthManager) GetProvider(name string) (OAuthProvider, error) {
	factory, exists := m.factories[name]
	if !exists {
		return nil, fmt.Errorf("provider factory not found: %s", name)
	}

	config, exists := m.configs[name]
	if !exists {
		return nil, fmt.Errorf("provider config not found: %s", name)
	}

	return factory.CreateProvider(config), nil
}
