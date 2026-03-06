/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package service

import (
	"costrict-keeper/cmd/root"

	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service operations (list/start/stop/restart etc.)",
	Long:  `Service operations (list/start/stop/restart etc.)`,
}

const serviceExample = `  # start service
  costrict service start codebase-indexer`

func init() {
	root.RootCmd.AddCommand(serviceCmd)

	serviceCmd.Example = serviceExample
}
