package component

import (
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Management components",
	Long:  `Management components, list, upgrade, remove, etc.`,
}

const componentExample = `  # Add task component
  # List components
  smc component list
  # List remote components
  smc component upgrade`

func init() {
	common.RootCmd.AddCommand(componentCmd)

	componentCmd.Example = componentExample
}
