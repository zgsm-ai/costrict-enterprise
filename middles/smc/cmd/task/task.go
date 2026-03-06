/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package task

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
)

// optTaskUUID is a global variable for task UUID used across multiple task subcommands
var optTaskUUID string

/**
 * Validate task/model name
 */
func CheckTaskName(modelName string, helpCmd string) error {
	if modelName == "" {
		return fmt.Errorf("missing model parameter, use '%s' for help", helpCmd)
	}
	return nil
}

// taskCmd represents the 'smc task' command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Task operations (view, stop etc.)",
	Long:  `'smc task' handles task operations like viewing and stopping`,
}

const taskExample = `  # Get task logs
  smc task logs 6e120fdf-6388-4615-bd3b-58aeb011a423
  # Get task status
  smc task list 6e120fdf-6388-4615-bd3b-58aeb011a423 -v
  # Stop task
  smc task stop 6e120fdf-6388-4615-bd3b-58aeb011a423
  # View task queue status
  smc task queue auto -v
  # View t4 resource pool details including usage and running tasks
  smc pool list t4 -v`

func init() {
	common.RootCmd.AddCommand(taskCmd)

	taskCmd.Example = taskExample
}
