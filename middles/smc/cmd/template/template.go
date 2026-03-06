/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
)

/**
 *	Check template name
 */
func CheckTemplateName(tplName string, helpCmd string) error {
	if tplName == "" {
		return fmt.Errorf("Parameter 'name' is missing, see '%s' for help", helpCmd)
	}
	return nil
}

// templateCmd represents the 'smc model' command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Operations for task template management",
	Long:  `'smc template' supports create/delete/list for task templates`,
}

const templateExample = `  # Create new task template
  smc template add codeview
  # Remove task template
  smc template rm codeview
  # List task templates
  smc template list`

func init() {
	common.RootCmd.AddCommand(templateCmd)

	templateCmd.Example = templateExample
}
