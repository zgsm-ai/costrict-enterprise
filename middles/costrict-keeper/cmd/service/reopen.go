package service

import (
	"encoding/json"
	"fmt"

	"costrict-keeper/internal/models"
	"costrict-keeper/internal/rpc"

	"github.com/spf13/cobra"
)

var reopenCmd = &cobra.Command{
	Use:   "reopen {service-name}",
	Short: "Reopen tunnel for specified service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		if serviceName == "" {
			fmt.Println("Must specify service name")
			return
		}

		reopenTunnel(serviceName)
	},
}

/**
 * Reopen tunnel via RPC connection to costrict server
 * @param {string} appName - Application name
 * @returns {bool} True if RPC call succeeded, false otherwise
 * @description
 * - Attempts to connect to costrict server via Unix socket
 * - Calls /costrict/api/v1/services/{appName}/reopen endpoint to reopen tunnel
 * - Handles connection errors and API response errors
 * - Returns success/failure status for fallback logic
 * @throws
 * - Connection establishment errors
 * - API request errors
 * - Response parsing errors
 */
func reopenTunnel(appName string) {
	rpcClient := rpc.NewHTTPClient(nil)
	resp, err := rpcClient.Post(fmt.Sprintf("/costrict/api/v1/services/%s/reopen", appName), nil)
	if err != nil {
		fmt.Printf("Failed to call costrict API: %v\n", err)
		return
	}
	if resp.Error != "" {
		fmt.Printf("Costrict API returned error(%d): %s\n", resp.StatusCode, resp.Error)
		return
	}
	var tun models.TunnelDetail
	if err := json.Unmarshal(resp.Body, &tun); err != nil {
		fmt.Printf("Failed to unmarshal tunnel instance: %v\n", err)
		return
	}

	// 成功反序列化，输出隧道信息
	fmt.Printf("Successfully reopened tunnel for %s\n", appName)
	fmt.Printf("  Name: %s\n", tun.Name)
	fmt.Printf("  Status: %s\n", tun.Status)
	fmt.Printf("  PID: %d\n", tun.Pid)
	fmt.Printf("  Created Time: %s\n", tun.CreatedTime.Format("2006-01-02 15:04:05"))
	if len(tun.Pairs) > 0 {
		fmt.Printf("  Local Port: %d -> Mapping Port: %d\n",
			tun.Pairs[0].LocalPort, tun.Pairs[0].MappingPort)
	}
}

func init() {
	serviceCmd.AddCommand(reopenCmd)
}
