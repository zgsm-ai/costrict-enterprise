/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package cmd

import (
	"github.com/iancoleman/orderedmap"
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/prompt"
	"github.com/zgsm-ai/smc/internal/utils"
)

/**
 *	Fields displayed in list format
 */
type Tool_Columns struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Module      string   `json:"module"`
	Type        string   `json:"type"`
	Method      string   `json:"method"`
	URL         string   `json:"url"`
	Supports    []string `json:"supports"`
	Description string   `json:"description"`
}

/**
 *	View pool information
 */
func toolList(toolId string, verbose bool) error {
	if err := common.InitRedisEnv(); err != nil {
		return err
	}
	if toolId != "" {
		tp, err := prompt.GetTool(toolId, verbose)
		if err != nil {
			return err
		}
		utils.PrintYaml(tp)
		return nil
	}
	tools, err := prompt.ListTools()
	if err != nil {
		return err
	}
	var dataList []*orderedmap.OrderedMap
	for k, p := range tools {
		row := Tool_Columns{}
		row.ID = k
		row.Module = p.Module
		row.Name = p.Name
		row.Description = p.Description
		row.Type = p.Type
		if p.Restful != nil {
			row.URL = p.Restful.Url
			row.Method = p.Restful.Method
		} else {
			row.URL = ""
			row.Method = ""
		}
		row.Supports = p.Supports

		recordMap, _ := utils.StructToOrderedMap(row)
		dataList = append(dataList, recordMap)
	}
	utils.PrintFormat(dataList)
	return nil
}

// toolListCmd represents the 'smc tool list' command
var toolListCmd = &cobra.Command{
	Use:   "list {tool | -i tool}",
	Short: "View tool definitions",
	Long:  `'smc tool list' shows tool definitions`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optToolId = args[0]
		}
		return toolList(optToolId, optToolVerbose)
	},
}

const toolListExample = `  # List tool definitions
  smc tool list`

var optToolVerbose bool

func init() {
	toolCmd.AddCommand(toolListCmd)
	toolListCmd.Flags().SortFlags = false
	toolListCmd.Example = toolListExample
	toolListCmd.Flags().StringVarP(&optToolId, "id", "i", "", "Tool definition ID")
	toolListCmd.Flags().BoolVarP(&optToolVerbose, "verbose", "v", false, "Show details")
}
