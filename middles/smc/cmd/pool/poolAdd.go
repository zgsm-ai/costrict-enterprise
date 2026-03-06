/*
Copyright © 2023 xmz <xmz@sangfor.com.cn>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/task"
)

func poolAdd() error {
	if err := common.InitTaskdEnv(); err != nil {
		return err
	}
	if optPool.Waiting == 0 {
		return fmt.Errorf("the parameter 'waiting' must be greater than 0")
	}
	if optPool.Running == 0 {
		return fmt.Errorf("the parameter 'running' must be greater than 0")
	}
	if optPool.PoolId == "" {
		return fmt.Errorf("the parameter 'pool' cannot be an empty string")
	}
	if optPool.Engine == "" {
		return fmt.Errorf("the parameter 'engine' cannot be an empty string")
	}
	if data, err := task.AddPool(common.Session, &optPool); err != nil {
		return err
	} else {
		fmt.Printf("%s", string(data))
	}
	return nil
}

// poolAddCmd represents the 'smc pool add' command
var poolAddCmd = &cobra.Command{
	Use:   "add {name | -n name} -e engine",
	Short: "Add task pool",
	Long:  `'smc pool add' adds a task pool`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optPool.PoolId = args[0]
		}
		return poolAdd()
	},
}

const poolAddExample = `  # Add task pool
  smc pool add codereview`

var optPool task.PoolBasic

func init() {
	poolCmd.AddCommand(poolAddCmd)
	poolAddCmd.Flags().SortFlags = false
	poolAddCmd.SilenceUsage = true
	poolAddCmd.Example = poolAddExample

	poolAddCmd.Flags().StringVarP(&optPool.PoolId, "name", "n", "", "Task pool name")
	poolAddCmd.Flags().StringVarP(&optPool.Engine, "engine", "e", "", "Task engine")
	poolAddCmd.Flags().StringVarP(&optPool.Description, "description", "d", "", "Task pool description")
	poolAddCmd.Flags().StringVarP(&optPool.Config, "config", "c", "", "Task pool configuration")
	poolAddCmd.Flags().IntVarP(&optPool.Running, "running", "C", 10, "Maximum concurrent tasks in pool")
	poolAddCmd.Flags().IntVarP(&optPool.Waiting, "waiting", "q", 20, "Maximum queued tasks in pool")
}
