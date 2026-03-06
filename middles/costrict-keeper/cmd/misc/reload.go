package client

import (
	"context"
	"fmt"

	"costrict-keeper/cmd/root"
	"costrict-keeper/internal/rpc"

	"github.com/spf13/cobra"
)

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload server configuration",
	Long:  `Reload server configuration by connecting to the costrict server and calling the reload API`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		reloadServerConfig(context.Background())
	},
}

/**
 * Reload server configuration via RPC connection to costrict server
 * @param {context.Context} ctx - Context for request cancellation and timeout
 * @returns {error} Returns error if reload fails, nil on success
 * @description
 * - Creates new RPC client instance
 * - Calls POST /costrict/api/v1/reload endpoint to reload configuration
 * - Handles connection errors and API response errors
 * - Outputs the reload status
 * - Used for remote configuration reload via costrict server
 * @throws
 * - Connection establishment errors
 * - API request errors
 * - Response parsing errors
 */
func reloadServerConfig(ctx context.Context) {
	rpcClient := rpc.NewHTTPClient(nil)

	// 调用 costrict 的 RESTful API POST 方法
	resp, err := rpcClient.Post("/costrict/api/v1/reload", nil)
	if err != nil {
		fmt.Printf("Failed to call costrict API: %v\n", err)
		return
	}
	if resp.Error != "" {
		fmt.Printf("Costrict API returned error(%d): %s\n", resp.StatusCode, resp.Error)
		return
	}
	fmt.Printf("Successfully reloaded server configuration, status code: %d\n", resp.StatusCode)
}

func init() {
	root.RootCmd.AddCommand(reloadCmd)
}
