package service

import (
	"fmt"

	"costrict-keeper/internal/rpc"

	"github.com/spf13/cobra"
)

var closeCmd = &cobra.Command{
	Use:   "close {service-name}",
	Short: "Close tunnel for specified serivce",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		if serviceName == "" {
			fmt.Println("Must specify service name")
			return
		}
		closeTunnel(serviceName)
	},
}

/**
 * Try to close tunnel via RPC connection to costrict server
 * @param {string} serviceName - Application name
 * @returns {void} No return value
 * @description
 * - Attempts to connect to costrict server via Unix socket
 * - Calls /costrict/api/v1/services/{service_name}/close endpoint to close tunnel
 * - Handles connection errors and API response errors
 * - Logs success/failure status messages
 * @throws
 * - Connection establishment errors
 * - API request errors
 * - Response parsing errors
 */
func closeTunnel(serviceName string) {
	rpcClient := rpc.NewHTTPClient(nil)
	// 尝试调用 costrict 的 RESTful API
	resp, err := rpcClient.Post(fmt.Sprintf("/costrict/api/v1/services/%s/close", serviceName), nil)
	if err != nil {
		fmt.Printf("Failed to call costrict API: %v\n", err)
		return
	}
	if resp.Error != "" {
		fmt.Printf("Costrict API returned error(%d): %s\n", resp.StatusCode, resp.Error)
		return
	}

	fmt.Printf("Tunnel [%s] stopped successfully\n", serviceName)
}

func init() {
	serviceCmd.AddCommand(closeCmd)
}
