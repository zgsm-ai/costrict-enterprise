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
 *	Columns to display for extensions
 */
type Extension_Columns struct {
	Key           string `json:"key"`
	Name          string `json:"name" description:"Extension name"`
	Publisher     string `json:"publisher" description:"Publisher name"`
	DisplayName   string `json:"displayName" description:"Extension display name"`
	Version       string `json:"version" description:"Extension version"`
	ExtensionType string `json:"extensionType" description:"Extension type"`
	License       string `json:"license" description:"Extension license"`
}

/**
 *	List extension pool information
 */
func extensionList(extensionId string, verbose bool) error {
	if err := common.InitRedisEnv(); err != nil {
		return err
	}
	if extensionId != "" {
		tp, err := prompt.GetExtension(extensionId)
		if err != nil {
			return err
		}
		utils.PrintYaml(tp)
		return nil
	}
	extensions, err := prompt.ListExtensions()
	if err != nil {
		return err
	}
	var dataList []*orderedmap.OrderedMap
	for k, p := range extensions {
		row := Extension_Columns{}
		row.Key = k
		row.Publisher = p.Publisher
		row.Name = p.Name
		row.ExtensionType = p.ExtensionType
		row.DisplayName = p.DisplayName
		// row.Description = p.Description
		// row.Icon = p.Icon
		row.License = p.License
		row.Version = p.Version

		recordMap, _ := utils.StructToOrderedMap(row)
		dataList = append(dataList, recordMap)
	}
	utils.PrintFormat(dataList)
	return nil
}

// extensionListCmd represents the 'smc extension list' command
var extensionListCmd = &cobra.Command{
	Use:   "list {extension | -i extension}",
	Short: "List extension definitions",
	Long:  `'smc extension list' lists extension definitions`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optExtensionId = args[0]
		}
		return extensionList(optExtensionId, optExtensionVerbose)
	},
}

const extensionListExample = `  # List extension definitions
  smc extension list`

var optExtensionVerbose bool

func init() {
	extensionCmd.AddCommand(extensionListCmd)
	extensionListCmd.Flags().SortFlags = false
	extensionListCmd.Example = extensionListExample
	extensionListCmd.Flags().StringVarP(&optExtensionId, "id", "i", "", "Extension ID")
	extensionListCmd.Flags().BoolVarP(&optExtensionVerbose, "verbose", "v", false, "Show details")
}
