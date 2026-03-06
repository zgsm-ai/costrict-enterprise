package service

import (
	"context"
	"costrict-keeper/internal/models"
	"costrict-keeper/internal/rpc"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart {service-name}",
	Short: "Restart service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		restartService(context.Background(), args[0])
	},
}

/**
 * Restart service via RPC client to costrict server
 * @param {context.Context} ctx - Context for request cancellation and timeout
 * @param {string} serviceName - Name of the service to restart
 * @returns {error} Returns error if RPC restart fails, nil on success
 * @description
 * - Creates RPC client with default configuration
 * - Attempts to connect to costrict server via Unix socket
 * - Calls restart API endpoint if connection succeeds
 * - Returns error if connection or API call fails
 * @throws
 * - RPC client creation errors
 * - Connection establishment errors
 * - API call errors
 * @example
 * err := restartService(context.Background(), "codebase-syncer")
 * if err != nil {
 *     fmt.Printf("RPC restart failed: %v", err)
 * }
 */
/**
 * Restart service via RPC client to costrict server
 * @param {context.Context} ctx - Context for request cancellation and timeout
 * @param {string} serviceName - Name of the service to restart
 * @returns {error} Returns error if RPC restart fails, nil on success
 * @description
 * - Creates RPC client with default configuration
 * - Attempts to connect to costrict server via HTTP
 * - Calls restart API endpoint if connection succeeds
 * - Parses and displays detailed service information after restart
 * - Returns error if connection or API call fails
 * @throws
 * - RPC client creation errors
 * - Connection establishment errors
 * - API call errors
 * - JSON parsing errors
 * @example
 * err := restartService(context.Background(), "codebase-syncer")
 * if err != nil {
 *     fmt.Printf("RPC restart failed: %v", err)
 * }
 */
func restartService(ctx context.Context, serviceName string) {
	rpcClient := rpc.NewHTTPClient(nil)
	resp, err := rpcClient.Post(fmt.Sprintf("/costrict/api/v1/services/%s/restart", serviceName), nil)
	if err != nil {
		fmt.Printf("failed to call costrict API: %w\n", err)
		return
	}
	if resp.Error != "" {
		fmt.Printf("Costrict API returned error(%d): %s\n", resp.StatusCode, resp.Error)
		return
	}

	var serviceDetail models.ServiceDetail
	if err := json.Unmarshal(resp.Body, &serviceDetail); err != nil {
		fmt.Printf("failed to unmarshal service detail: %w\n", err)
		return
	}

	// 成功重启服务，显示服务详细信息
	fmt.Printf("Successfully restarted service '%s'\n", serviceName)
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
	serviceCmd.AddCommand(restartCmd)
}
