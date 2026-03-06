package service

import (
	"costrict-keeper/internal/rpc"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop {service-name}",
	Short: "Stop service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := stopService(args[0]); err != nil {
			fmt.Println(err)
		}
	},
}

/**
 * Stop service by name
 * @param {string} serviceName - Name of the service to stop
 * @returns {error} Returns error if service stop fails, nil on success
 * @description
 * - Creates new service manager instance
 * - Attempts to stop the specified service
 * - Prints success message if service stops successfully
 * @throws
 * - Service stop failure errors
 * @example
 * err := stopService("codebase-syncer")
 * if err != nil {
 *     logger.Fatal(err)
 * }
 */
func stopService(serviceName string) error {
	rpcClient := rpc.NewHTTPClient(nil)
	defer rpcClient.Close()

	apiPath := fmt.Sprintf("/costrict/api/v1/services/%s/stop", serviceName)
	resp, err := rpcClient.Post(apiPath, nil)
	if err == nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			fmt.Printf("Failed to stop service '%s': %+v\n", serviceName, resp.Body)
			return os.ErrInvalid
		}
		fmt.Printf("Service '%s' has been stopped\n", serviceName)
		return nil
	}
	fmt.Printf("Service '%s' stop failed\n", serviceName)
	return nil
}

func init() {
	serviceCmd.AddCommand(stopCmd)
}
