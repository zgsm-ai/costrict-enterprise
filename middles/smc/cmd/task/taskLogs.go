/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package task

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/task"
)

func taskLogs() error {
	if err := common.InitTaskdEnv(); err != nil {
		return err
	}
	var err error
	var arg task.TaskLogsArgs
	arg.Follow = optTaskFollow
	arg.Tail = int64(optTaskTail)
	arg.Entity = optEntityName
	arg.Timestamps = optTaskTimestamp
	if optTaskUUID != "" {
		err = task.GetTaskLogs(common.Session, optTaskUUID, &arg)
	} else {
		err = fmt.Errorf("missing parameters, please specify task uuid or runid")
	}
	return err
}

// taskLogsCmd represents the 'smc task logs' command
var taskLogsCmd = &cobra.Command{
	Use:   "logs {UUID | -i UUID}",
	Short: "View task logs",
	Long:  `'smc task logs' displays task logs on the platform`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optTaskUUID = args[0]
		}
		return taskLogs()
	},
}

const taskLogsExample = `
# View task logs
smc task logs a8664ea43aa94bd28081943a5827ef78
# View worker-1's logs
smc task logs a8664ea43aa94bd28081943a5827ef78 -p worker-1
# Follow worker-1's logs
smc task logs a8664ea43aa94bd28081943a5827ef78 -p worker-1 -f
`

var optEntityName string
var optTaskTimestamp bool
var optTaskFollow bool
var optTaskTail int

func init() {
	taskCmd.AddCommand(taskLogsCmd)
	taskLogsCmd.Flags().SortFlags = false
	taskLogsCmd.Example = taskLogsExample
	taskLogsCmd.Flags().StringVarP(&optTaskUUID, "uuid", "i", "", "Task UUID")
	taskLogsCmd.Flags().StringVarP(&optEntityName, "entity", "e", "", "Task entity name")
	taskLogsCmd.Flags().IntVarP(&optTaskTail, "tail", "t", 100, "Number of log lines")
	taskLogsCmd.Flags().BoolVarP(&optTaskFollow, "follow", "f", false, "Follow log output")
	taskLogsCmd.Flags().BoolVarP(&optTaskTimestamp, "timestamps", "s", false, "Show timestamps")
}
