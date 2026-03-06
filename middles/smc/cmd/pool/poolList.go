/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package cmd

import (
	"fmt"

	"github.com/iancoleman/orderedmap"
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/task"
	"github.com/zgsm-ai/smc/internal/utils"
)

/**
 *	Fields displayed in list format
 */
type Pool_Columns struct {
	PoolId      string
	Engine      string
	MaxWaiting  string
	MaxRunning  string
	Waiting     string
	Running     string
	Description string
}

/**
 *	View pool information
 */
func poolList(poolName string, verbose bool) error {
	if err := common.InitTaskdEnv(); err != nil {
		return err
	}
	if poolName != "" {
		tp, err := task.GetPool(common.Session, poolName, verbose)
		if err != nil {
			return err
		}
		utils.PrintYaml(tp)
		return nil
	}
	pools, err := task.ListPools(common.Session)
	if err != nil {
		return err
	}
	var dataList []*orderedmap.OrderedMap
	for _, p := range pools {
		row := Pool_Columns{}
		row.PoolId = p.PoolId
		row.Engine = p.Engine
		row.Description = p.Description
		row.MaxRunning = fmt.Sprint(p.MaxRunning)
		row.MaxWaiting = fmt.Sprint(p.MaxWaiting)
		row.Waiting = fmt.Sprint(p.Waiting)
		row.Running = fmt.Sprint(p.Running)

		recordMap, _ := utils.StructToOrderedMap(row)
		dataList = append(dataList, recordMap)
	}
	utils.PrintFormat(dataList)
	return nil
}

// poolListCmd represents the 'smc pool list' command
var poolListCmd = &cobra.Command{
	Use:   "list {pool | -p pool}",
	Short: "View task pool status",
	Long:  `'smc pool list' can view task pool status`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optPoolName = args[0]
		}
		return poolList(optPoolName, optPoolVerbose)
	},
}

const taskPoolExample = `  # List resource pools
  smc pool list`

var optPoolName string
var optPoolVerbose bool

func init() {
	poolCmd.AddCommand(poolListCmd)
	poolListCmd.Flags().SortFlags = false
	poolListCmd.Example = taskPoolExample
	poolListCmd.Flags().StringVarP(&optPoolName, "pool", "p", "", "Pool name")
	poolListCmd.Flags().BoolVarP(&optPoolVerbose, "verbose", "v", false, "Show details")
}
