/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	common "github.com/zgsm-ai/smc/cmd/common"
	env "github.com/zgsm-ai/smc/internal/env"
)

func configSet() error {
	if err := common.InitCommonEnv(); err != nil {
		return err
	}
	for _, kv := range envKvs {
		if kv.Value == "" {
			continue
		}
		if err := env.EnableModify(kv.Name); err != nil {
			return err
		}
		fmt.Printf("%s = %s, old value: %s\n", kv.Name, kv.Value, env.GetEnv(kv.Name))
		if err := env.SetEnv(kv.Name, kv.Value); err != nil {
			return err
		}
	}
	if optKey != "" {
		if err := env.EnableModify(optKey); err != nil {
			return err
		}
		fmt.Printf("%s = %s, old value: %s\n", optKey, optValue, env.GetEnv(optKey))
		if err := env.SetEnv(optKey, optValue); err != nil {
			return err
		}
	}
	return nil
}

// configSetCmd represents the 'smc config set' command
var configSetCmd = &cobra.Command{
	Use:   "set {-k key -v value | --optkey optvalue}...",
	Short: "Set smc configurations",
	Long:  `'smc config set' updates smc configuration values`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := configSet(); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	envCmd.AddCommand(configSetCmd)

	configSetCmd.Example = `  # Set envDebug to 'Dbg'
  smc config set --envDebug Dbg`

	configSetCmd.Flags().SortFlags = false
	configSetCmd.Flags().StringVarP(&optKey, "key", "k", "", "Configuration name")
	configSetCmd.Flags().StringVarP(&optValue, "value", "v", "", "Configuration value")

	env.InitEnvs()
	cnt := 0
	env.VisitAll(func(name, alias, comment, defVal string, value env.OptValue) error {
		cnt++
		return nil
	})
	envKvs = make([]common.OptKeyValue, cnt)
	idx := 0
	env.VisitAll(func(name, alias, comment, defVal string, value env.OptValue) error {
		envKvs[idx].Name = name
		configSetCmd.Flags().StringVarP(&envKvs[idx].Value, alias, "", "", comment+" ["+name+"]")
		idx++
		return nil
	})
}
