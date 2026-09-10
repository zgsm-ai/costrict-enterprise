package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/zgsm-ai/oidc-auth/internal/config"
	"github.com/zgsm-ai/oidc-auth/internal/handler"
	"github.com/zgsm-ai/oidc-auth/internal/providers"
	"github.com/zgsm-ai/oidc-auth/internal/repository"
	"github.com/zgsm-ai/oidc-auth/internal/service"
	github "github.com/zgsm-ai/oidc-auth/internal/sync"
	"github.com/zgsm-ai/oidc-auth/internal/usercenter"
	"github.com/zgsm-ai/oidc-auth/pkg/log"
	"github.com/zgsm-ai/oidc-auth/pkg/utils"
)

var (
	cfgFile      string
	globalConfig *config.AppConfig
	client       *http.Client
	githubClient *http.Client
)

var rootCmd = &cobra.Command{
	Use:   "oidc-auth",
	Short: "OIDC Authentication Server",
	Long:  `OIDC Authentication Server using Casdoor for authentication and authorization.`,
}

func initLogger(cfg *config.LogConfig) error {
	log.InitLogger(&log.Config{
		Level:    cfg.Level,
		Filename: cfg.Filename,
		MaxSize:  cfg.MaxSize,
		MaxAge:   cfg.MaxAge,
		Compress: cfg.Compress,
	})
	return nil
}

// initDatabase initializes database connection
func initDatabase(cfg *config.DatabaseConfig) error {
	dbCfg := repository.DBConfig(*cfg)
	if _, err := repository.InitGlobalDatabase(&dbCfg); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	return nil
}

func initializeAllConfigurations(cfgFile string) (*config.AppConfig, error) {
	cfg, err := config.InitConfig(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	if cfg == nil {
		return nil, fmt.Errorf("failed to initialize config: nil config")
	}

	utils.SetGlobalConfig(cfg)

	// RSA key loading is unconditional and lazy (sync.Once); fail fast at
	// startup so a missing or broken encrypt.privateKey surfaces now rather
	// than at the first token sign/verify call.
	if _, err := utils.GetEncryptKeyManager(); err != nil {
		return nil, fmt.Errorf("failed to load encryption keys: %w", err)
	}

	if err := initLogger(&cfg.Log); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	if err := initDatabase(&cfg.Database); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return cfg, nil
}

func newHTTPClient(cfg *config.HTTPClientConfig, proxyURL string) (*http.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("HTTP client configuration is required")
	}

	proxy := http.ProxyFromEnvironment
	if proxyURL != "" {
		parsedProxy, err := url.Parse(proxyURL)
		if err != nil || parsedProxy.Host == "" || !supportedProxyScheme(parsedProxy.Scheme) {
			return nil, fmt.Errorf("invalid GitHub proxy URL")
		}
		proxy = http.ProxyURL(parsedProxy)
	}

	transport := &http.Transport{
		Proxy: proxy,
		DialContext: (&net.Dialer{
			Timeout:   cfg.DialTimeout,
			KeepAlive: cfg.KeepAlive,
		}).DialContext,
		TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       cfg.IdleConnTimeout,
	}
	return &http.Client{Transport: transport, Timeout: cfg.Timeout}, nil
}

func supportedProxyScheme(scheme string) bool {
	switch strings.ToLower(scheme) {
	case "http", "https", "socks5", "socks5h":
		return true
	default:
		return false
	}
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the OIDC authentication server",
	PreRun: func(cmd *cobra.Command, args []string) {
		var err error
		globalConfig, err = initializeAllConfigurations(cfgFile)
		if err != nil {
			log.Fatal(nil, "Failed to initialize config: %v", err)
		}
		if globalConfig.GithubConfig.Webhook.Enabled &&
			(globalConfig.GithubConfig.Webhook.Secret == "" || globalConfig.GithubConfig.Owner == "" || globalConfig.GithubConfig.Repo == "") {
			log.Fatal(nil, "GitHub webhook owner, repo, and secret must be configured when the webhook is enabled")
		}

		client, err = newHTTPClient(globalConfig.Server.HTTP, "")
		if err != nil {
			log.Fatal(nil, "Failed to initialize HTTP client: %v", err)
		}
		githubClient, err = newHTTPClient(globalConfig.Server.HTTP, globalConfig.GithubConfig.ProxyURL)
		if err != nil {
			log.Fatal(nil, "Failed to initialize GitHub HTTP client: %v", err)
		}

		httpClient := client
		smsc := service.GetSMSCfg(&globalConfig.SMS)
		if smsc == nil {
			log.Fatal(nil, "Failed to initialize SMS service")
		}
		smsc.HTTPClient = httpClient

		// Initialize quota service
		globalConfig.QuotaManager.HTTPClient = httpClient
		service.InitQuotaService(&globalConfig.QuotaManager)
		providerCfg := make(map[string]*providers.ProviderConfig)
		for name, p := range globalConfig.Providers {
			providerCfg[name] = &providers.ProviderConfig{
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				BaseURL:      p.BaseURL,
				Client:       httpClient,
				InternalURL:  p.InternalURL,
			}
		}
		err = providers.InitializeProviders(providerCfg)
		if err != nil {
			log.Fatal(err, "Failed to initialize providers")
		}

		// Initialize cs-user client (fail-closed: without it the login chain
		// cannot establish the identity trust boundary).
		if err := usercenter.InitClient(globalConfig.UserCenter.BaseURL, globalConfig.UserCenter.InternalToken, httpClient, globalConfig.UserCenter.Timeout); err != nil {
			log.Fatal(err, "Failed to initialize usercenter client")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		syncStar := github.SyncStar{
			Enabled:       globalConfig.GithubConfig.Enabled,
			PersonalToken: globalConfig.GithubConfig.PersonalToken,
			Owner:         globalConfig.GithubConfig.Owner,
			Repo:          globalConfig.GithubConfig.Repo,
			Interval:      globalConfig.GithubConfig.Interval,
			HTTPClient:    githubClient,
		}
		github.Owner, github.Repo = syncStar.Owner, syncStar.Repo
		go syncStar.StarSyncTimer(ctx)

		go func() {
			log.Info(nil, "Starting server...")
			server := handler.Server{
				ServerPort:           globalConfig.Server.ServerPort,
				BaseURL:              globalConfig.Server.BaseURL,
				WebBaseURL:           globalConfig.Server.WebBaseURL,
				HTTPClient:           client,
				IsPrivate:            globalConfig.Server.IsPrivate,
				RedirectURL:          globalConfig.Redirect.Uris,
				GitHubWebhookEnabled: globalConfig.GithubConfig.Webhook.Enabled,
				GitHubWebhookSecret:  globalConfig.GithubConfig.Webhook.Secret,
				GitHubOwner:          globalConfig.GithubConfig.Owner,
				GitHubRepo:           globalConfig.GithubConfig.Repo,
			}
			if err := server.StartServer(); err != nil {
				log.Error(nil, "Server error: %v", err)
				cancel()
			}
		}()

		<-ctx.Done()
		log.Info(nil, "Shutting down server...")
	},
}

func init() {
	serveCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.AddCommand(serveCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
