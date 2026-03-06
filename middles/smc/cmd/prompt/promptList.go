/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package cmd

import (
	"github.com/iancoleman/orderedmap"
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/prompt"
	"github.com/zgsm-ai/smc/internal/utils"
)

/**
 *	Fields displayed in list format
 */
type Prompt_Columns struct {
	Key      string
	Name     string
	Prompt   string
	Supports []string
}

/**
 *	View pool information
 */
func promptList(promptId string, verbose bool) error {
	if err := common.InitRedisEnv(); err != nil {
		return err
	}
	if promptId != "" {
		tp, err := prompt.GetPrompt(promptId, verbose)
		if err != nil {
			return err
		}
		utils.PrintYaml(tp)
		return nil
	}
	prompts, err := prompt.ListPrompts()
	if err != nil {
		return err
	}
	var dataList []*orderedmap.OrderedMap
	for k, p := range prompts {
		row := Prompt_Columns{}
		row.Key = k
		row.Name = p.Name
		row.Prompt = p.Prompt
		row.Supports = p.Supports

		recordMap, _ := utils.StructToOrderedMap(row)
		dataList = append(dataList, recordMap)
	}
	utils.PrintFormat(dataList)
	return nil
}

// promptListCmd represents the 'smc prompt list' command
var promptListCmd = &cobra.Command{
	Use:   "list {prompt | -i prompt}",
	Short: "View prompt templates",
	Long:  `'smc prompt list' shows prompt template details`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optPromptId = args[0]
		}
		return promptList(optPromptId, optPromptVerbose)
	},
}

const promptListExample = `  # List all prompt template names
  smc prompt list`

var optPromptVerbose bool

func init() {
	promptCmd.AddCommand(promptListCmd)
	promptListCmd.Flags().SortFlags = false
	promptListCmd.Example = promptListExample
	promptListCmd.Flags().StringVarP(&optPromptId, "id", "i", "", "Prompt template ID")
	promptListCmd.Flags().BoolVarP(&optPromptVerbose, "verbose", "v", false, "Show details")
}
