/*
Copyright © 2023 xmz <xmz@sangfor.com.cn>
*/
package task

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/task"
)

func taskStop() error {
	if err := common.InitTaskdEnv(); err != nil {
		return err
	}
	var err error
	if optTaskUUID != "" {
		err = task.StopTask(common.Session, optTaskUUID)
	} else {
		fmt.Println("Missing parameters, please specify task uuid or instance runid")
	}
	return err
}

// taskStopCmd represents the 'smc task stop' command
var taskStopCmd = &cobra.Command{
	Use:   "stop {UUID | -i UUID}",
	Short: "Stop task",
	Long:  `'smc task stop' stops a running task`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optTaskUUID = args[0]
		}
		return taskStop()
	},
}

const taskStopExample = `  # Stop a task by uuid or instance ID (stopping only the specified instance)
  smc task stop 65c462ec-e011-4b12-ab21-a28fb25bdc30
  smc task stop 1`

func init() {
	taskCmd.AddCommand(taskStopCmd)
	taskStopCmd.Flags().SortFlags = false
	taskStopCmd.Example = taskStopExample

	taskStopCmd.Flags().StringVarP(&optTaskUUID, "uuid", "i", "", "Specify task UUID")
}
