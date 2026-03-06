/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/iancoleman/orderedmap"
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/prompt"
	"github.com/zgsm-ai/smc/internal/utils"
)

func printVariable(val interface{}, format string) {
	if format == "plain" {
		fmt.Println(val)
	} else if format == "yaml" {
		utils.PrintYaml(val)
	} else if format == "json" {
		data, err := json.MarshalIndent(val, "", "  ")
		if err != nil {
			panic("json marshal error")
		}
		fmt.Println(string(data))
	} else {
		panic("unsupported format")
	}
}

/**
 *	Show pool information
 */
func environList(environId string) error {
	if err := common.InitRedisEnv(); err != nil {
		return err
	}
	if environId != "" {
		tp, err := prompt.GetEnviron(environId)
		if err != nil {
			return err
		}
		printVariable(tp, optEnvironFormat)
		return nil
	}
	environs, err := prompt.ListEnvirons()
	if err != nil {
		return err
	}
	var dataList []*orderedmap.OrderedMap
	for k, p := range environs {
		type Environ_Columns struct {
			Key   string
			Value string
		}
		row := Environ_Columns{}
		row.Key = k
		row.Value = fmt.Sprintf("%+v", p)

		recordMap, _ := utils.StructToOrderedMap(row)
		dataList = append(dataList, recordMap)
	}
	utils.PrintFormat(dataList)
	return nil
}

// variableListCmd represents the 'smc variable list' command
var variableListCmd = &cobra.Command{
	Use:   "list {id | -i id}",
	Short: "Check shared variables status",
	Long:  `'smc variable list' checks shared variables status`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optEnvironId = args[0]
		}
		return environList(optEnvironId)
	},
}

const variableListExample = `  # List shared variables
  smc variable list`

func init() {
	variableCmd.AddCommand(variableListCmd)
	variableListCmd.Flags().SortFlags = false
	variableListCmd.Example = variableListExample
	variableListCmd.Flags().StringVarP(&optEnvironId, "id", "i", "", "Shared variable ID")
	variableListCmd.Flags().StringVarP(&optEnvironFormat, "format", "f", "plain", "Output format: supports plain, yaml, json")
}
