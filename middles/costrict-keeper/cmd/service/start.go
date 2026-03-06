package service

import (
	"encoding/json"
	"fmt"
	"time"

	"costrict-keeper/internal/models"
	"costrict-keeper/internal/rpc"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start {service-name}",
	Short: "Start service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		if serviceName == "" {
			fmt.Println("Must specify service name")
			return
		}

		startService(serviceName)
	},
}

/**
 * Start service via RPC connection to costrict server
 * @param {string} serviceName - Service name to start
 * @returns {void} No return value, outputs results directly or exits on error
 * @description
 * - Attempts to connect to costrict server via Unix socket
 * - Calls /costrict/api/v1/services/{serviceName}/start endpoint to start service
 * - Handles connection errors and API response errors
 * - Displays success message if service starts successfully
 * @throws
 * - Connection establishment errors
 * - API request errors
 * - Response parsing errors
 * @example
 * startService("codebase-syncer")
 */
func startService(serviceName string) {
	rpcClient := rpc.NewHTTPClient(nil)
	resp, err := rpcClient.Post(fmt.Sprintf("/costrict/api/v1/services/%s/start", serviceName), nil)
	if err != nil {
		fmt.Printf("Failed to call costrict API: %v\n", err)
		return
	}
	if resp.Error != "" {
		fmt.Printf("Costrict API returned error: %s\n", resp.Error)
		return
	}
	var serviceDetail models.ServiceDetail
	if err := json.Unmarshal(resp.Body, &serviceDetail); err != nil {
		fmt.Printf("Failed to unmarshal service detail: %v\n", err)
		return
	}

	// 成功启动服务，显示服务详细信息
	fmt.Printf("Successfully started service '%s'\n", serviceName)
	fmt.Printf("  Name: %s\n", serviceDetail.Name)
	fmt.Printf("  Status: %s\n", serviceDetail.Status)
	fmt.Printf("  PID: %d\n", serviceDetail.Pid)
	if serviceDetail.Port > 0 {
		fmt.Printf("  Port: %d\n", serviceDetail.Port)
	}
	if serviceDetail.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, serviceDetail.StartTime)
		if err == nil {
			fmt.Printf("  Start Time: %s\n", startTime.Format("2006-01-02 15:04:05"))
		}
	}
	if serviceDetail.Tunnel != nil {
		fmt.Printf("  Tunnel: %s\n", serviceDetail.Tunnel.Status)
		for _, pair := range serviceDetail.Tunnel.Pairs {
			fmt.Printf("    Local Port: %d -> Mapping Port: %d\n", pair.LocalPort, pair.MappingPort)
		}
	}
}

func init() {
	serviceCmd.AddCommand(startCmd)
}
