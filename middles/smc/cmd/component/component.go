package component

import (
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/env"
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

var mirrorUrl string

func init() {
	common.RootCmd.AddCommand(componentCmd)

	componentCmd.Example = componentExample
	componentCmd.PersistentFlags().StringVarP(&mirrorUrl, "mirror", "m", "", "mirror site")
}

func getUpgradeUrl() string {
	if mirrorUrl != "" {
		return mirrorUrl + "/costrict"
	}
	return env.BaseUrl + "/costrict"
}
