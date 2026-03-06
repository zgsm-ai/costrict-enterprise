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

func promptAdd() error {
	if err := common.InitRedisEnv(); err != nil {
		return err
	}
	if optPromptDefFile == "" {
		return fmt.Errorf("file notexist")
	}
	content, err := os.ReadFile(optPromptDefFile)
	if err != nil {
		return err
	}
	promptDef := prompt.Prompt{}
	if err := json.Unmarshal(content, &promptDef); err != nil {
		return err
	}
	if err := prompt.AddPrompt(optPromptId, &promptDef); err != nil {
		return err
	} else {
		utils.PrintYaml(&promptDef)
	}
	return nil
}

// promptAddCmd represents the 'smc prompt add' command
var promptAddCmd = &cobra.Command{
	Use:   "add {id | -i id} -d definition-json-file",
	Short: "Add prompt template",
	Long:  `'smc prompt add' adds a prompt template`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optPromptId = args[0]
		}
		return promptAdd()
	},
}

const promptAddExample = `  # Add a prompt template
  smc prompt add codereview.memory.leak --definition examples/prompt.json`

var optPromptDefFile string

func init() {
	promptCmd.AddCommand(promptAddCmd)
	promptAddCmd.Flags().SortFlags = false
	promptAddCmd.SilenceUsage = true
	promptAddCmd.Example = promptAddExample

	promptAddCmd.Flags().StringVarP(&optPromptId, "id", "i", "", "Prompt template ID")
	promptAddCmd.Flags().StringVarP(&optPromptDefFile, "definition", "d", "prompt.json", "Prompt template definition file in JSON format")
}
