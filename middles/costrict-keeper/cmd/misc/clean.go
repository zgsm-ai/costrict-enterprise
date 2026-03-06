package client

import (
	"costrict-keeper/cmd/root"
	"costrict-keeper/internal/config"
	"costrict-keeper/internal/env"
	"costrict-keeper/internal/utils"
	"costrict-keeper/services"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up all services, tunnels, processes and cache",
	Long:  `Stop all services, terminate all tunnels, kill specified processes and clean up .costrict cache directory`,
	Run: func(cmd *cobra.Command, args []string) {
		cleanAll()
	},
}

/**
 * Clean up all services, tunnels, processes and cache
 * @returns {error} Returns error if any cleanup step fails, nil on success
 * @description
 * - Stops all running services
 * - Closes all active tunnels
 * - Kills processes with names: costrict, codebase-indexer, cotun
 * - Cleans up .costrict cache directory
 * @throws
 * - Service stop errors
 * - Tunnel close errors
 * - Process kill errors
 * - Cache cleanup errors
 */
func cleanAll() error {
	// 1. 杀掉还在运行的costrict程序
	fmt.Println("Starting cleanup process...")
	utils.KillSpecifiedProcess(services.COSTRICT_NAME)
	// 2. 杀死所有组件/服务的进程
	if err := config.LoadSpec(); err == nil {
		spec := config.Spec()
		targetProcesses := []string{}
		for _, svc := range spec.Components {
			targetProcesses = append(targetProcesses, svc.Name)
		}
		utils.KillSpecifiedProcesses(targetProcesses)
	}
	fmt.Println("The remaining processes have been successfully cleaned up.")
	// 3. 清理.costrict目录下的cache目录
	cleanCacheDirectory()
	fmt.Println("Cache directory cleaned successfully")
	fmt.Println("Clean completed successfully")
	return nil
}

/**
 * Clean up .costrict cache directory
 * @returns {error} Returns error if cache cleanup fails, nil on success
 * @description
 * - Gets .costrict directory path from config
 * - Constructs cache directory path
 * - Removes cache directory and all its contents
 * @throws
 * - Directory path construction errors
 * - Directory removal errors
 */
func cleanCacheDirectory() {
	fmt.Println("Cleaning up cache directory...")

	costrictDir := env.CostrictDir
	if costrictDir == "" {
		fmt.Println("Failed to get .costrict directory path")
		return
	}

	cacheDir := filepath.Join(costrictDir, "cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		fmt.Printf("Cache directory '%s' does not exist, skipping cleanup\n", cacheDir)
		return
	}

	// 删除cache/services目录及其所有内容
	servicesDir := filepath.Join(cacheDir, "services")
	if err := os.RemoveAll(servicesDir); err != nil {
		fmt.Printf("Failed to remove cache directory %s: %v\n", servicesDir, err)
	} else {
		fmt.Printf("Successfully removed cache directory: %s\n", servicesDir)
	}

	// 删除cache/tunnels目录及其所有内容
	tunnelsDir := filepath.Join(cacheDir, "tunnels")
	if err := os.RemoveAll(tunnelsDir); err != nil {
		fmt.Printf("Failed to remove cache directory %s: %v\n", tunnelsDir, err)
	} else {
		fmt.Printf("Successfully removed cache directory: %s\n", tunnelsDir)
	}

	// 删除run目录及其所有内容
	runDir := filepath.Join(env.CostrictDir, "run")
	if err := os.RemoveAll(runDir); err != nil {
		fmt.Printf("Failed to remove run directory %s: %v\n", runDir, err)
	} else {
		fmt.Printf("Successfully removed run directory: %s\n", runDir)
	}
}

func init() {
	root.RootCmd.AddCommand(cleanCmd)
}
