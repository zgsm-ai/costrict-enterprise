/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
)

/**
 *	Check pool name
 */
func CheckPoolName(tplName string, helpCmd string) error {
	if tplName == "" {
		return fmt.Errorf("Parameter 'name' is missing, use '%s' for more help", helpCmd)
	}
	return nil
}

// poolCmd represents the 'smc pool' command
var poolCmd = &cobra.Command{
	Use:   "pool",
	Short: "Operations for managing task pools (create/delete/list)",
	Long:  `'smc pool' can be used to manage task pools (create/delete/list)`,
}

const poolExample = `  # Add task pool
  smc pool add rpc
  # Remove task pool
  smc pool rm rpc
  # List task pools
  smc pool list`

func init() {
	common.RootCmd.AddCommand(poolCmd)

	poolCmd.Example = poolExample
}
