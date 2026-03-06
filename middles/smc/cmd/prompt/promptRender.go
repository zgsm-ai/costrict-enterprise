/*
Copyright © 2023 xmz <xmz@sangfor.com.cn>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/prompt"
)

func promptRender() error {
	if err := common.InitPromptEnv(); err != nil {
		return err
	}
	var args map[string]interface{}
	if optPromptArgs != "" {
		if err := json.Unmarshal([]byte(optPromptArgs), &args); err != nil {
			return err
		}
	}
	if data, err := prompt.RenderPrompt(common.Session, optPromptId, args); err != nil {
		return err
	} else {
		fmt.Println(string(data))
	}
	return nil
}

// promptRenderCmd represents the 'smc prompt add' command
var promptRenderCmd = &cobra.Command{
	Use:   "render {id | -i id}",
	Short: "Get rendered result of specified prompt template",
	Long:  `'smc prompt render' renders prompt template with arguments`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optPromptId = args[0]
		}
		return promptRender()
	},
}

const promptRenderExample = `  # Render codereview.memory.leak template
  smc prompt render codereview.memory.leak --args '{"file": "examples/test.go"}'`

func init() {
	promptCmd.AddCommand(promptRenderCmd)
	promptRenderCmd.Flags().SortFlags = false
	promptRenderCmd.SilenceUsage = true
	promptRenderCmd.Example = promptRenderExample

	promptRenderCmd.Flags().StringVarP(&optPromptId, "id", "i", "", "Prompt template ID")
	promptRenderCmd.Flags().StringVarP(&optPromptArgs, "args", "v", "", "Arguments for template rendering")
}
