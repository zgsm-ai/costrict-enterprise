/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package component

import (
	"costrict-keeper/cmd/root"

	"github.com/spf13/cobra"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Component operations (list/upgrade/remove etc.)",
	Long:  `Component operations (list/upgrade/remove etc.)`,
}

const componentExample = `  # list component
  costrict component list
  costrict component upgrade codebase-indexer
  costrict component remove codebase-indexer
  costrict component upgrade -n codebase-indexer
  costrict component remove -n codebase-indexer`

func init() {
	root.RootCmd.AddCommand(componentCmd)

	componentCmd.Example = componentExample
}
