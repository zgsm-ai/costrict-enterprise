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

func promptChat() error {
	if err := common.InitPromptEnv(); err != nil {
		return err
	}
	if optPromptArgs == "" {
		optPromptArgs = "{}"
	}
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(optPromptArgs), &args); err != nil {
		return err
	}
	if data, err := prompt.ChatWithPrompt(common.Session, optPromptId, optPromptModel, args); err != nil {
		return err
	} else {
		fmt.Println(string(data))
	}
	return nil
}

// promptChatCmd represents the 'smc prompt add' command
var promptChatCmd = &cobra.Command{
	Use:   "chat {id | -i id} -m model",
	Short: "Chat with LLM using specified prompt template",
	Long:  `'smc prompt chat' starts chat with LLM using specified prompt template`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optPromptId = args[0]
		}
		return promptChat()
	},
}

const promptChatExample = `  # Use codereview.memory.leak template with default model
  smc prompt chat codereview.memory.leak --args '{"file": "examples/test.go"}'`

var optPromptArgs string
var optPromptModel string

func init() {
	promptCmd.AddCommand(promptChatCmd)
	promptChatCmd.Flags().SortFlags = false
	promptChatCmd.SilenceUsage = true
	promptChatCmd.Example = promptChatExample

	promptChatCmd.Flags().StringVarP(&optPromptId, "id", "i", "", "Prompt template ID")
	promptChatCmd.Flags().StringVarP(&optPromptModel, "model", "m", "deepseek-v3", "LLM model for conversation")
	promptChatCmd.Flags().StringVarP(&optPromptArgs, "args", "v", "", "Arguments for template rendering")
}
