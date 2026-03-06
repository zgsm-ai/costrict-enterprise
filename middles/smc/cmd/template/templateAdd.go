/*
Copyright © 2023 xmz <xmz@sangfor.com.cn>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/task"
	"github.com/zgsm-ai/smc/internal/utils"
)

func templateAdd() error {
	if err := common.InitTaskdEnv(); err != nil {
		return err
	}
	if optSchemaFile != "" {
		schemaData, err := os.ReadFile(optSchemaFile)
		if err != nil {
			return err
		}
		optTemplate.Schema = string(schemaData)
	}
	if err := task.AddTemplate(common.Session, optTemplate); err != nil {
		return err
	}
	utils.PrintYaml(optTemplate)
	return nil
}

// templateAddCmd represents the 'smc template add' command
var templateAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a task template",
	Long:  `'smc template add' creates a new task template`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optTemplate.Name = args[0]
		}
		return templateAdd()
	},
}

const templateAddExample = `  # View task definition
  smc template add codereview -e rpc`

var optTemplate task.TemplateMetadata
var optSchemaFile string

func init() {
	templateCmd.AddCommand(templateAddCmd)
	templateAddCmd.Flags().SortFlags = false
	templateAddCmd.Example = templateAddExample

	templateAddCmd.Flags().StringVarP(&optTemplate.Name, "name", "n", "", "Task type name")
	templateAddCmd.Flags().StringVarP(&optTemplate.Title, "title", "t", "", "Title")
	templateAddCmd.Flags().StringVarP(&optTemplate.Engine, "engine", "e", "", "Execution engine")
	templateAddCmd.Flags().StringVarP(&optTemplate.Extra, "extra", "E", "", "Extra parameters")
	templateAddCmd.Flags().StringVarP(&optTemplate.Schema, "schema", "s", "", "Task metadata(template)")
	templateAddCmd.Flags().StringVarP(&optSchemaFile, "schema-file", "f", "", "Task metadata file")
}
