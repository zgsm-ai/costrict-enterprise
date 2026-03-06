package client

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"costrict-keeper/cmd/root"
	"costrict-keeper/internal/config"
	"costrict-keeper/internal/env"
	"costrict-keeper/services"
)

var (
	optUploadFile      string
	optUploadDirectory string
	logService         *services.LogService
)

func init() {
	root.RootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().SortFlags = false
	uploadCmd.Flags().StringVarP(&optUploadFile, "file", "f", "", "Log file path")
	uploadCmd.Flags().StringVarP(&optUploadDirectory, "directory", "d", "", "Log directory path")
}

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload logs to the cloud",
	Run: func(cmd *cobra.Command, args []string) {
		if optUploadFile == "" && optUploadDirectory == "" {
			optUploadDirectory = filepath.Join(env.CostrictDir, "logs")
		}
		logService = services.NewLogService()

		if optUploadFile != "" {
			err := logService.UploadFile(optUploadFile)
			if err != nil {
				fmt.Printf("Failed to upload file '%s' to '%s': %v\n", optUploadFile, config.Cloud().LogUrl, err)
				return
			}
			fmt.Printf("Upload successful: %s\n", optUploadFile)
		} else {
			err := logService.UploadDirectory(optUploadDirectory)
			if err != nil {
				fmt.Printf("Failed to upload directory '%s' to '%s': %v\n", optUploadDirectory, config.Cloud().LogUrl, err)
				return
			}
			fmt.Printf("Upload successful: %s\n", optUploadDirectory)
		}
		auth := config.GetAuthConfig()
		fmt.Printf("Cloud URL: %s\n", auth.BaseUrl)
		fmt.Printf("Client ID: %s\n", auth.MachineID)
		fmt.Printf("User ID:   %s\n", auth.ID)
	},
}
