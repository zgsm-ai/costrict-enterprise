/*
Copyright © 2023 xmz <xmz@sangfor.com.cn>
*/
package task

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/env"
	"github.com/zgsm-ai/smc/internal/task"
)

/**
 *	Parse environment variables into map
 */
func parseKVS(name string, kvs []string) (map[string]string, error) {
	envs := make(map[string]string)
	for _, kvStr := range kvs {
		kv := strings.SplitN(kvStr, "=", 2)
		if len(kv) != 2 {
			return envs, fmt.Errorf("Failed to parse %s parameter, format should be: key=value (multiple allowed)", name)
		}
		envs[kv[0]] = kv[1]
	}
	return envs, nil
}

func taskSubmit() error {
	if err := common.InitTaskdEnv(); err != nil {
		return err
	}
	tags, err := parseKVS("tags", optTaskTags)
	if err != nil {
		return err
	}
	data, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	optTask.Tags = string(data)
	optTask.Callback = env.Callback
	if env.Listen != "" {
		go RunHttpServer()
		if data, err = task.StartTask(common.Session, &optTask); err != nil {
			return err
		}
		req := <-CallbackChan
		fmt.Printf("%s\n", string(data))
		fmt.Printf("%+v\n", req)
	} else {
		if data, err = task.StartTask(common.Session, &optTask); err != nil {
			return err
		}
		fmt.Printf("%s\n", string(data))
	}
	return nil
}

// taskSubmitCmd represents the 'smc task submit' command
var taskSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a task",
	Long: `Usage:
	 'smc task submit' submits a task to specified task pool
  `,
	Args: cobra.MaximumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return taskSubmit()
	},
}

const taskSubmitExample = `  # Submit a new task`

var optTask task.TaskMetadata
var optTaskTags []string

func init() {
	taskCmd.AddCommand(taskSubmitCmd)
	taskSubmitCmd.Flags().SortFlags = false
	taskSubmitCmd.Example = taskSubmitExample

	taskSubmitCmd.Flags().StringVarP(&optTask.Name, "name", "n", "", "Task name (optional), will auto generate unique name if not specified")
	taskSubmitCmd.Flags().StringVarP(&optTask.Template, "template", "t", "", "Task template")
	taskSubmitCmd.Flags().StringVarP(&optTask.Extra, "extra", "e", "", "Task extra parameters (for overriding same-name settings in template)")
	taskSubmitCmd.Flags().StringVarP(&optTask.Args, "args", "a", "", "Task user arguments")
	taskSubmitCmd.Flags().StringVarP(&optTask.Pool, "pool", "p", "", "Task pool name (optional)")
	taskSubmitCmd.Flags().StringVarP(&optTask.Project, "project", "P", "", "Project name")
	taskSubmitCmd.Flags().StringVarP(&optTask.Namespace, "user", "u", "", "Username (optional), will use current login username if not specified")
	taskSubmitCmd.Flags().StringSliceVarP(&optTaskTags, "tags", "T", []string{}, "Task tags, support: gpumem=xG,gpu=a800(vGPU,rtx4090,v100,a30,h100) etc. (optional)")
}
