package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/zgsm-ai/client-manager/controllers"
	"github.com/zgsm-ai/client-manager/internal"
	"github.com/zgsm-ai/client-manager/router"
	"github.com/zgsm-ai/client-manager/services"
)

var SoftwareVer = ""
var BuildTime = ""
var BuildTag = ""
var BuildCommitId = ""

func PrintVersions() {
	fmt.Printf("Version %s\n", SoftwareVer)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Build Tag: %s\n", BuildTag)
	fmt.Printf("Build Commit ID: %s\n", BuildCommitId)
}

// @title Client Manager API
// @version 1.0
// @description This is a client manager API server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "client-manager",
	Short: "Client Manager API Server",
	Long:  `Client Manager is a RESTful API server for managing client configurations, feedback, and logs.`,
	Run: func(cmd *cobra.Command, args []string) {
		PrintVersions()
		// Load configuration
		if err := internal.LoadConfig(internal.AppConfig.ConfigPath); err != nil {
			fmt.Printf("Failed to load configuration: %v\n", err)
			os.Exit(1)
		}
		// Apply command line overrides
		internal.ApplyConfig()

		// Initialize application
		app, err := services.InitializeApp()
		if err != nil {
			fmt.Printf("Failed to initialize application: %v\n", err)
			os.Exit(1)
		}

		// Initialize controllers
		logController := controllers.NewLogController(app.Logger, app.LogService)

		// Create Gin engine
		r := gin.Default()

		// Setup all routes
		router.SetupRoutes(r, logController, app.Logger)

		// Start server
		if err := services.StartServer(r, app.Logger); err != nil {
			app.Logger.Fatalf("Failed to start server: %v", err)
		}
		gracefulShutdown(app)
	},
}

func init() {
	internal.InitFlags(rootCmd)
}

// gracefulShutdown sets up graceful shutdown handlers
/**
* Setup graceful shutdown handlers
* @param {*services.AppContext} app - Application context containing database connection
* @description
* - Sets up signal handlers for SIGINT and SIGTERM
* - Closes database connection gracefully
* - Logs shutdown process
 */
func gracefulShutdown(app *services.AppContext) {
	app.Logger.Info("Shutting down application...")

	// Close database connection
	if err := internal.CloseDB(); err != nil {
		app.Logger.WithError(err).Error("Failed to close database connection")
	} else {
		app.Logger.Info("Database connection closed successfully")
	}

	app.Logger.Info("Application shutdown completed")

}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
