package pkg

import (
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
)

// packageCmd represents the 'smc package' command
var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Management packages",
	Long:  `Management packages, such as package creation, merging, and upgrading, etc.`,
}

const packageExample = `  # Add task package
  smc package build rpc
  # Merge package version
  smc package index rpc
  # List packages
  smc package list
  # List remote packages
  smc package remote`

func init() {
	common.RootCmd.AddCommand(packageCmd)

	packageCmd.Example = packageExample
}
