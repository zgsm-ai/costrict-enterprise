/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package cmd

import (
	"github.com/spf13/cobra"
	common "github.com/zgsm-ai/smc/cmd/common"
)

/**
 * Config command for managing smc configurations
 */
var envCmd = &cobra.Command{
	Use:   "config",
	Short: "View and modify smc configurations",
	Long:  `'smc config' manages configurations used by smc`,
}

/**
 * Alias option for config commands
 */
var optAlias string

/**
 * Key option for config set commands
 */
var optKey string

/**
 * Value option for config set commands
 */
var optValue string

/**
 * Key-value pairs for config operations
 */
var envKvs []common.OptKeyValue

/**
 * Initialize config command with flags and examples
 */
func init() {
	common.RootCmd.AddCommand(envCmd)

	envCmd.Example = `  smc config
  # List all configurations
  smc config list
  # Set configuration
  smc config set KEY VALUE`
}
