package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zgsm-ai/oidc-auth/internal/middleware"
	"github.com/zgsm-ai/oidc-auth/internal/providers"
	"github.com/zgsm-ai/oidc-auth/pkg/log"
)

type Server struct {
	ServerPort           string
	BaseURL              string
	WebBaseURL           string
	HTTPClient           *http.Client
	IsPrivate            bool
	RedirectURL          map[string]string
	GitHubWebhookEnabled bool
	GitHubWebhookSecret  string
	GitHubOwner          string
	GitHubRepo           string
	updateGithubStar     func(context.Context, string, string, bool) (int64, error)
}

// webRedirectBase returns the origin used for post-login browser redirects
// (the web frontend, e.g. http://127.0.0.1:9527 in local dev). Falls back to
// the OAuth provider's public endpoint when the dedicated web base URL is not
// configured, so existing deployments keep their current behavior.
func (s *Server) webRedirectBase(providerInstance providers.OAuthProvider) string {
	if s.WebBaseURL != "" {
		return s.WebBaseURL
	}
	return providerInstance.GetEndpoint(false)
}

type ParameterCarrier struct {
	TokenHash     string `json:"token_hash"`
	Provider      string `json:"provider"`
	Platform      string `json:"platform"`
	State         string `form:"state" binding:"required"`
	MachineCode   string `form:"machine_code"`
	UriScheme     string `form:"uri_scheme"`
	PluginVersion string `form:"plugin_version"`
	VscodeVersion string `form:"vscode_version"`
}

func (s *Server) SetupRouter(r *gin.Engine) {
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RequestLogger())

	pluginOauthServer := r.Group("/oidc-auth/api/v1/plugin",
		middleware.SetPlatform("plugin"),
	)
	{
		pluginOauthServer.GET("login", s.loginHandler)
		pluginOauthServer.GET("login/callback", s.callbackHandler)
		pluginOauthServer.GET("login/token", tokenHandler)
		pluginOauthServer.GET("login/logout", logoutHandler)
		pluginOauthServer.GET("login/status", statusHandler)
	}
	webOauthServer := r.Group("/oidc-auth/api/v1/manager",
		middleware.SetPlatform("web"),
	)
	{
		webOauthServer.GET("token", getTokenByHash)
		webOauthServer.GET("bind/account", s.bindAccount)
		webOauthServer.GET("bind/account/callback", s.bindAccountCallback)
		webOauthServer.GET("userinfo", s.userInfoHandler)
		webOauthServer.GET("login", s.webLoginHandler) // web login entry point and redirect to custom service
		webOauthServer.GET("login/callback", s.webLoginCallbackHandler)
		webOauthServer.GET("login/callback/:service", s.webLoginCallbackHandler) // with service param, custom redirect
		webOauthServer.GET("invite-code", s.getUserInviteCodeHandler)
	}
	r.POST("/oidc-auth/api/v1/send/sms", s.SMSHandler)
	if s.GitHubWebhookEnabled {
		r.POST("/oidc-auth/api/v1/webhooks/github", s.githubWebhookHandler)
	}
	health := r.Group("/health")
	{
		health.GET("ready", readinessHandler)
	}
}

func (s *Server) StartServer() error {
	r := gin.Default()
	s.SetupRouter(r)

	port := ":" + s.ServerPort
	log.Info(nil, "Starting server on port %s", port)

	if err := r.Run(port); err != nil {
		log.Error(nil, "Server failed to start: %v", err)
		return err
	}
	return nil
}
