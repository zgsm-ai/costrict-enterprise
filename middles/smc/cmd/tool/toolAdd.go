/*
Copyright © 2023 xmz <xmz@sangfor.com.cn>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/prompt"
	"github.com/zgsm-ai/smc/internal/utils"
)

func toolAdd() error {
	if err := common.InitRedisEnv(); err != nil {
		return err
	}
	if optToolDefFile == "" {
		return fmt.Errorf("file notexist")
	}
	content, err := os.ReadFile(optToolDefFile)
	if err != nil {
		return err
	}
	toolDef := prompt.Tool{}
	if err := json.Unmarshal(content, &toolDef); err != nil {
		return err
	}
	if err := prompt.AddTool(optToolId, &toolDef); err != nil {
		return err
	} else {
		utils.PrintYaml(&toolDef)
	}
	return nil
}

// toolAddCmd represents the 'smc tool add' command
var toolAddCmd = &cobra.Command{
	Use:   "add {id | -i id} -d definition-json-file",
	Short: "Add tool definition",
	Long:  `'smc tool add' command adds tool definitions`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optToolId = args[0]
		}
		return toolAdd()
	},
}

const toolAddExample = `  # Add tool definition
  smc tool add codebase.lookup_ref -d examples/tool.json`

var optToolDefFile string

func init() {
	toolCmd.AddCommand(toolAddCmd)
	toolAddCmd.Flags().SortFlags = false
	toolAddCmd.SilenceUsage = true
	toolAddCmd.Example = toolAddExample

	toolAddCmd.Flags().StringVarP(&optToolId, "id", "i", "", "Tool definition ID")
	toolAddCmd.Flags().StringVarP(&optToolDefFile, "definition", "d", "tool.json", "Tool definition file (JSON format)")
}
