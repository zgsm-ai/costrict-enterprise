/*
Copyright © 2023 xmz <xmz@sangfor.com.cn>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/task"
)

func templateRm() error {
	if err := common.InitTaskdEnv(); err != nil {
		return err
	}
	if err := task.RemoveTemplate(common.Session, optTemplateName); err != nil {
		return err
	}
	return nil
}

// templateRmCmd represents the 'smc template rm' command
var templateRmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove task template",
	Long:  `'smc template rm' removes a task template`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optTemplateName = args[0]
		}
		return templateRm()
	},
}

const templateRmExample = `  # Remove task template
  smc template rm codereview`

var optTemplateName string

func init() {
	templateCmd.AddCommand(templateRmCmd)
	templateRmCmd.Flags().SortFlags = false
	templateRmCmd.Example = templateRmExample

	templateRmCmd.Flags().StringVarP(&optTemplateName, "name", "n", "", "Task type name")
}
