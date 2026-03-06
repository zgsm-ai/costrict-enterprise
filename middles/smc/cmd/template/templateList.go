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

func printTemplate(t *task.TemplateMetadata) error {
	extra := t.Extra
	schema := t.Schema

	if extra != "" {
		t.Extra = "..."
	}
	if schema != "" {
		t.Schema = "..."
	}

	if err := utils.PrintYaml(t); err != nil {
		return err
	}
	if optTemplateList.Verbose {
		if extra != "" {
			fmt.Printf("----------------extra----------------\n%s\n", common.GetPrettyJson(extra))
		}
		if schema != "" {
			fmt.Printf("----------------schema---------------\n%s\n", schema)
		}
	}
	return nil
}

func showTemplate() error {
	md, err := task.GetTemplate(common.Session, optTemplateList.Name)
	if err != nil {
		return err
	}
	printTemplate(&md)
	return nil
}

/**
 *	Fields displayed in list format
 */
type TemplateColumn struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Engine      string `json:"engine"`
	Schema      string `json:"schema"`
	Extra       string `json:"extra"`
	CreatedBy   string `json:"created_by"`
	CreateTime  string `json:"create_time"`
}

func templateList() error {
	if err := common.InitTaskdEnv(); err != nil {
		return err
	}
	if optTemplateList.Name != "" {
		return showTemplate()
	}
	tpls, err := task.ListTemplates(common.Session, &optTemplateList)
	if err != nil {
		return err
	}
	if len(tpls) == 0 {
		return nil
	} else if len(tpls) == 1 {
		return printTemplate(&tpls[0])
	}
	var dataList []*orderedmap.OrderedMap
	for _, v := range tpls {
		col := TemplateColumn{}
		col.Name = v.Name
		col.Title = v.Title
		col.Schema = v.Schema
		col.Description = v.Description
		col.Engine = v.Engine
		col.Extra = v.Extra

		om, err := utils.StructToOrderedMap(col)
		if err != nil {
			return err
		}
		dataList = append(dataList, om)
	}
	return utils.PrintFormat(dataList)
}

// templateListCmd represents the 'smc template list' command
var templateListCmd = &cobra.Command{
	Use:   "list {name | -n name}",
	Short: "View task templates",
	Long:  `'smc template list' shows defined task templates`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optTemplateList.Name = args[0]
		}
		return templateList()
	},
}

const templateListExample = `
# View task list
smc template list a8664ea43aa94bd28081943a5827ef78
smc template list -p <project>
`

var optTemplateList task.ListTemplatesArgs

func init() {
	templateCmd.AddCommand(templateListCmd)
	templateListCmd.Flags().SortFlags = false
	templateListCmd.Example = templateListExample
	templateListCmd.Flags().StringVarP(&optTemplateList.Name, "name", "n", "", "Task template name")
	templateListCmd.Flags().StringVarP(&optTemplateList.Title, "title", "t", "", "Task template display title")
	templateListCmd.Flags().StringVarP(&optTemplateList.Engine, "engine", "e", "", "Task engine")
	templateListCmd.Flags().BoolVarP(&optTemplateList.Verbose, "verbose", "v", false, "Show detailed information including extra, yamlContent, endLog")
	templateListCmd.Flags().IntVarP(&optTemplateList.Page, "page", "g", 1, "Starting page number")
	templateListCmd.Flags().IntVarP(&optTemplateList.PageSize, "pageSize", "m", 10, "Number of task records to display")
}
