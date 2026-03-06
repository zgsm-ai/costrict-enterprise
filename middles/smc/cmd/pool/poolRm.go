/*
Copyright © 2023 xmz <xmz@sangfor.com.cn>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/task"
)

func poolRm() error {
	if err := common.InitTaskdEnv(); err != nil {
		return err
	}
	if err := task.RemovePool(common.Session, optPoolName); err != nil {
		return err
	}
	return nil
}

// poolRmCmd represents the 'smc pool rm' command
var poolRmCmd = &cobra.Command{
	Use:   "rm {name | -n name}",
	Short: "Remove task pool",
	Long:  `'smc pool rm' removes a task pool`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optPoolName = args[0]
		}
		return poolRm()
	},
}

const poolRmExample = `  # Remove task pool
  smc pool rm codereview`

func init() {
	poolCmd.AddCommand(poolRmCmd)
	poolRmCmd.Flags().SortFlags = false
	poolRmCmd.SilenceUsage = true
	poolRmCmd.Example = poolRmExample

	poolRmCmd.Flags().StringVarP(&optPoolName, "name", "n", "", "Task pool name")
}
