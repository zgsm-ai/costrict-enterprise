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

func taskTags() error {
	if err := common.InitTaskdEnv(); err != nil {
		return err
	}
	if optTaskUUID == "" {
		return fmt.Errorf("Missing parameters: please specify task UUID or run ID of the task instance")
	}

	var err error
	var result map[string]string
	if optTagsKey != "" {
		result, err = task.SetTaskTags(common.Session, optTaskUUID, optTagsKey, optTagsValue)
	} else {
		result, err = task.GetTaskTags(common.Session, optTaskUUID)
	}
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", result)
	return nil
}

// taskTagsCmd represents the 'smc task tags' command
var taskTagsCmd = &cobra.Command{
	Use:   "tags {UUID | -i UUID}",
	Short: "Add tags to task instances",
	Long:  `'smc task tags' adds tags to task instances`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optTaskUUID = args[0]
		}
		return taskTags()
	},
}

const taskTagsExample = `  # Tag the last instance of a task
  smc task tags 65c462ec-e011-4b12-ab21-a28fb25bdc30 --key sla --value immediate
  smc task tags 1`

var optTagsKey string
var optTagsValue string

func init() {
	taskCmd.AddCommand(taskTagsCmd)
	taskTagsCmd.Flags().SortFlags = false
	taskTagsCmd.Example = taskTagsExample

	taskTagsCmd.Flags().StringVarP(&optTaskUUID, "uuid", "i", "", "Task UUID to specify")
	taskTagsCmd.Flags().StringVarP(&optTagsKey, "key", "k", "", "Tag key")
	taskTagsCmd.Flags().StringVarP(&optTagsValue, "value", "v", "", "Tag value")
}
