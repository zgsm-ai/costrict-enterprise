package client

import (
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Management clients",
	Long:  `Management clients, such as: logs, metrics, etc.`,
}

const clientExample = `  # 
  smc client logs
  smc client logs -c xxxx`

func init() {
	common.RootCmd.AddCommand(clientCmd)

	clientCmd.Example = clientExample
}
