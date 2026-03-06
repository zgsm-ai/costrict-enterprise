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

func taskStatus() error {
	if err := common.InitTaskdEnv(); err != nil {
		return err
	}
	if optTaskUUID != "" {
		status, err := task.GetTaskStatus(common.Session, optTaskUUID)
		if err != nil {
			return err
		}
		fmt.Println(string(status))
	} else {
		fmt.Println("Missing parameters, please specify task uuid or runid")
	}
	return nil
}

// taskStatusCmd represents the 'smc task status' command
var taskStatusCmd = &cobra.Command{
	Use:   "status {UUID | -i UUID}",
	Short: "View task status",
	Long:  `'smc task status' shows task status on the platform`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optTaskUUID = args[0]
		}
		return taskStatus()
	},
}

const taskStatusExample = `
# View task status with uuid ccddeeff
smc task status ccddeeff
`

func init() {
	taskCmd.AddCommand(taskStatusCmd)
	taskStatusCmd.Flags().SortFlags = false
	taskStatusCmd.Example = taskStatusExample
	taskStatusCmd.Flags().StringVarP(&optTaskUUID, "uuid", "i", "", "Specify task UUID")
}
