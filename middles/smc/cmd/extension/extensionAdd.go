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

func extensionAdd() error {
	if err := common.InitRedisEnv(); err != nil {
		return err
	}
	if optExtensionDefFile == "" {
		return fmt.Errorf("file notexist")
	}
	content, err := os.ReadFile(optExtensionDefFile)
	if err != nil {
		return err
	}
	ext := prompt.PromptExtension{}
	if err := json.Unmarshal(content, &ext); err != nil {
		return err
	}

	if err := prompt.AddExtension(optExtensionId, &ext); err != nil {
		return err
	} else {
		utils.PrintYaml(&ext)
	}
	return nil
}

// extensionAddCmd represents the 'smc extension add' command
var extensionAddCmd = &cobra.Command{
	Use:   "add {id | -i id} -d definition-json-file",
	Short: "Add extension",
	Long:  `'smc extension add' adds an extension`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optExtensionId = args[0]
		}
		return extensionAdd()
	},
}

const extensionAddExample = `  # Add extension
  smc extension add "agent.codereview" -d package.json`

var optExtensionDefFile string

func init() {
	extensionCmd.AddCommand(extensionAddCmd)
	extensionAddCmd.Flags().SortFlags = false
	extensionAddCmd.SilenceUsage = true
	extensionAddCmd.Example = extensionAddExample

	extensionAddCmd.Flags().StringVarP(&optExtensionId, "id", "i", "", "Extension ID")
	extensionAddCmd.Flags().StringVarP(&optExtensionDefFile, "definition", "d", "package.json", "Extension definition file in JSON format")
}
