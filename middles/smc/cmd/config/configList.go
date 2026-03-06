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

func configList(envName string) error {
	if err := common.InitCommonEnv(); err != nil {
		return err
	}
	if envName == "" && optAlias == "" && optKey == "" {
		env.VisitAll(func(name, alias, comment, defVal string, value env.OptValue) error {
			fmt.Printf("%-16s = %-30s # %-11s: %s (default %s)\n", name, value.String(), alias, comment, defVal)
			return nil
		})
		return nil
	}
	found := false
	env.VisitAll(func(name, alias, comment, defVal string, value env.OptValue) error {
		if optKey == name {
			fmt.Printf("%s\n", value.String())
			found = true
		} else if optAlias == alias {
			fmt.Printf("%s\n", value.String())
			found = true
		} else if envName == name || envName == alias {
			fmt.Printf("%s\n", value.String())
			found = true
		}
		return nil
	})
	if found {
		return nil
	}
	return fmt.Errorf("not found")
}

// configListCmd represents the 'smc config list' command
var configListCmd = &cobra.Command{
	Use:   "list [-k key | -a alias | key-or-alias]",
	Short: "View smc configurations",
	Long:  `'smc config list' views smc configurations`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		envName := ""
		if len(args) == 1 {
			envName = args[0]
		}
		if err := configList(envName); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	envCmd.AddCommand(configListCmd)

	configListCmd.Example = `  # View all smc configurations
  smc config list
  smc config list --key SHENMA_DEBUG
  smc config list --alias envDebug
  smc config list envDebug
  smc config list SHENMA_DEBUG
  `

	configListCmd.Flags().SortFlags = false
	configListCmd.Flags().StringVarP(&optKey, "key", "k", "", "Configuration name")
	configListCmd.Flags().StringVarP(&optAlias, "alias", "a", "", "Configuration alias")
}
