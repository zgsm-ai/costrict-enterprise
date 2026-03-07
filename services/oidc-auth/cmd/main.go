package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/zgsm-ai/oidc-auth/internal/config"
	"github.com/zgsm-ai/oidc-auth/internal/handler"
	"github.com/zgsm-ai/oidc-auth/internal/providers"
	"github.com/zgsm-ai/oidc-auth/internal/repository"
	"github.com/zgsm-ai/oidc-auth/internal/service"
	github "github.com/zgsm-ai/oidc-auth/internal/sync"
	"github.com/zgsm-ai/oidc-auth/pkg/log"
	"github.com/zgsm-ai/oidc-auth/pkg/utils"
)

var (
	cfgFile      string
	globalConfig *config.AppConfig
	initOnce     sync.Once
	client       *http.Client
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

	if err := initLogger(&cfg.Log); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	if err := initDatabase(&cfg.Database); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return cfg, nil
}

func initHTTPClient(cfg *config.HTTPClientConfig) *http.Client {
	initOnce.Do(func() {
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   cfg.DialTimeout,
				KeepAlive: cfg.KeepAlive,
			}).DialContext,
			TLSHandshakeTimeout: cfg.TLSHandshakeTimeout,

			ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,

			MaxIdleConns:        cfg.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
			IdleConnTimeout:     cfg.IdleConnTimeout,
		}
		client = &http.Client{Transport: transport, Timeout: cfg.Timeout}
	})
	return client
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
		httpClient := initHTTPClient(globalConfig.Server.HTTP)
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
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		syncStar := github.SyncStar(globalConfig.GithubConfig)
		syncStar.HTTPClient = initHTTPClient(nil)
		github.Owner, github.Repo = syncStar.Owner, syncStar.Repo
		go syncStar.StarSyncTimer(ctx)

		go func() {
			log.Info(nil, "Starting server...")
			server := handler.Server{
				ServerPort:  globalConfig.Server.ServerPort,
				BaseURL:     globalConfig.Server.BaseURL,
				HTTPClient:  initHTTPClient(nil),
				IsPrivate:   globalConfig.Server.IsPrivate,
				RedirectURL: globalConfig.Redirect.Uris,
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
