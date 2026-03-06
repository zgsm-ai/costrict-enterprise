/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
)

// variableCmd represents the 'smc variable' command
var variableCmd = &cobra.Command{
	Use:   "variable",
	Short: "Operations on shared variable definitions in REDIS",
	Long:  `'smc variable' can be used to manage shared variable definitions in REDIS`,
}

const variableExample = `  # List shared variable definitions in REDIS
  smc variable list`

var optEnvironId string
var optEnvironValue string
var optEnvironType string
var optEnvironFormat string

func init() {
	common.RootCmd.AddCommand(variableCmd)

	variableCmd.Example = variableExample
}
