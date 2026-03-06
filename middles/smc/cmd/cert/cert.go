/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package cert

import (
	"github.com/spf13/cobra"
	common "github.com/zgsm-ai/smc/cmd/common"
)

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Certified the data",
	Long:  `Certified the data to prevent it from being tampered with.`,
}

/**
 * Initialize config command with flags and examples
 */
func init() {
	common.RootCmd.AddCommand(certCmd)

	certCmd.Example = `  smc cert genkey
  smc cert sign -k costrict-private.pem -t ./shenma
  smc cert sum -f ./shenma`
}
