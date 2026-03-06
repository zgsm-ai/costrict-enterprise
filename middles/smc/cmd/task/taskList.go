/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package task

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/iancoleman/orderedmap"
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/task"
	"github.com/zgsm-ai/smc/internal/utils"
)

func printPrettyLog(log string) error {
	var val interface{}
	if err := json.Unmarshal([]byte(log), &val); err != nil {
		fmt.Println(log)
		return err
	}
	return utils.PrintYaml(val)
}

func printTask(tm *task.TaskMetadata) error {
	extra := tm.Extra
	tm.Extra = "..."

	if err := utils.PrintYaml(tm); err != nil {
		return err
	}
	if optList.Verbose {
		fmt.Printf("----------------extra----------------\n%s\n", common.GetPrettyJson(extra))
	}
	return nil
}

func showTask() error {
	tm, err := task.GetTask(common.Session, optTaskUUID, optList.Verbose)
	if err != nil {
		return err
	}
	extra := tm.Extra
	yamlContent := tm.YamlContent
	endLog := tm.EndLog

	if extra != "" {
		tm.Extra = "..."
	}
	if yamlContent != "" {
		tm.YamlContent = "..."
	}
	if endLog != "" {
		tm.EndLog = "..."
	}

	if err := utils.PrintYaml(tm); err != nil {
		return err
	}
	if optList.Verbose {
		if extra != "" {
			fmt.Printf("----------------extra----------------\n%s\n", common.GetPrettyJson(extra))
		}
		if yamlContent != "" {
			fmt.Printf("-------------yamlContent-------------\n%s\n", yamlContent)
		}
		if endLog != "" {
			fmt.Printf("-------------endLog------------------\n")
			printPrettyLog(endLog)
		}
	}
	return nil
}

/**
 * Task columns for listing display
 */
type TaskColumn struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Template   string `json:"template"`
	Project    string `json:"project"`
	Pool       string `json:"pool"`
	Namespace  string `json:"namespace"`
	CreatedBy  string `json:"created_by"`
	CreateTime string `json:"create_time"`
	Age        string `json:"age,omitempty"`
}

func taskList() error {
	if err := common.InitTaskdEnv(); err != nil {
		return err
	}
	if optTaskUUID != "" {
		return showTask()
	}
	optList.Sort = "id DESC"
	tasks, err := task.ListTasks(common.Session, &optList)
	if err != nil {
		return err
	}
	if len(tasks.List) == 0 {
		return nil
	} else if len(tasks.List) == 1 {
		return printTask(&tasks.List[0])
	}
	var dataList []*orderedmap.OrderedMap
	for _, v := range tasks.List {
		age, err := utils.FormatDuration(time.RFC3339, v.CreateTime, v.EndTime)
		if err != nil {
			return err
		}
		col := TaskColumn{}
		col.Name = v.Name
		col.CreateTime = v.CreateTime
		col.Status = v.Status
		col.Project = v.Project
		col.Pool = v.Pool
		col.Template = v.Template
		col.UUID = v.UUID
		col.Namespace = v.Namespace
		col.CreatedBy = v.CreatedBy
		col.Age = age

		om, err := utils.StructToOrderedMap(col)
		if err != nil {
			return err
		}
		dataList = append(dataList, om)
	}
	return utils.PrintFormat(dataList)
}

// taskListCmd represents the 'smc task list' command
var taskListCmd = &cobra.Command{
	Use:   "list {UUID | -i UUID}",
	Short: "List tasks",
	Long:  `'smc task list' displays tasks on the platform`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optTaskUUID = args[0]
		}
		return taskList()
	},
}

const taskListExample = `  # List tasks
  smc task list a8664ea43aa94bd28081943a5827ef78
  smc task list -p <project>`

var optList task.ListTasksArgs

func init() {
	taskCmd.AddCommand(taskListCmd)
	taskListCmd.Flags().SortFlags = false
	taskListCmd.Example = taskListExample
	taskListCmd.Flags().StringVarP(&optList.Uuid, "uuid", "i", "", "Task UUID to query")
	taskListCmd.Flags().StringVarP(&optList.Name, "name", "m", "", "Task name to query")
	taskListCmd.Flags().StringVarP(&optList.Namespace, "namespace", "u", "", "Task namespace")
	taskListCmd.Flags().StringVarP(&optList.Project, "project", "p", "", "Project name")
	taskListCmd.Flags().StringVarP(&optList.Status, "status", "s", "", "Task status")
	taskListCmd.Flags().StringVarP(&optList.Template, "type", "t", "", "Task type name")
	taskListCmd.Flags().StringVar(&optList.Pool, "pool", "", "Resource pool name")
	taskListCmd.Flags().BoolVarP(&optList.Verbose, "verbose", "v", false, "Show details (extra,yamlContent,endLog)")
	taskListCmd.Flags().IntVarP(&optList.Page, "page", "g", 1, "Start page number")
	taskListCmd.Flags().IntVarP(&optList.PageSize, "pageSize", "n", 10, "Number of records to display")
}
