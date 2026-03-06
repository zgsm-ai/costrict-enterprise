/*
Copyright © 2023 xmz <xmz@sangfor.com.cn>
*/
package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/prompt"
)

func variableSet() error {
	if err := common.InitRedisEnv(); err != nil {
		return err
	}
	if optEnvironType == "string" {
		if err := prompt.SetEnviron(optEnvironId, optEnvironValue); err != nil {
			return err
		}
	} else if optEnvironType == "object" {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(optEnvironValue), &obj); err != nil {
			return err
		}
		if err := prompt.SetEnviron(optEnvironId, obj); err != nil {
			return err
		}
	} else if optEnvironType == "array" {
		var arr []interface{}
		if err := json.Unmarshal([]byte(optEnvironValue), &arr); err != nil {
			return err
		}
		if err := prompt.SetEnviron(optEnvironId, arr); err != nil {
			return err
		}
	} else if optEnvironType == "number" {
		var num json.Number
		if err := json.Unmarshal([]byte(optEnvironValue), &num); err != nil {
			return err
		}
		if err := prompt.SetEnviron(optEnvironId, num); err != nil {
			return err
		}
	} else if optEnvironType == "bool" {
		var bol bool
		if err := json.Unmarshal([]byte(optEnvironValue), &bol); err != nil {
			return err
		}
		if err := prompt.SetEnviron(optEnvironId, bol); err != nil {
			return err
		}
	} else {
		panic("unknown type")
	}
	return nil
}

// variableSetCmd represents the 'smc variable set' command
var variableSetCmd = &cobra.Command{
	Use:   "set {id | -i id} -v value [-t type]",
	Short: "Add shared variables",
	Long:  `'smc variable set' adds shared variables`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optEnvironId = args[0]
		}
		return variableSet()
	},
}

const variableSetExample = `  # Set shared variable
  smc variable set 'codereview.supports' -v 'cpp'`

func init() {
	variableCmd.AddCommand(variableSetCmd)
	variableSetCmd.Flags().SortFlags = false
	variableSetCmd.SilenceUsage = true
	variableSetCmd.Example = variableSetExample

	variableSetCmd.Flags().StringVarP(&optEnvironId, "id", "i", "", "KEY")
	variableSetCmd.Flags().StringVarP(&optEnvironValue, "value", "v", "", "VALUE")
	variableSetCmd.Flags().StringVarP(&optEnvironType, "type", "t", "string", "type: string,array,object,number,bool")
}
