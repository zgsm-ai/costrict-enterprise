/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package cert

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/internal/utils"
)

func md5sum() error {
	_, md5str, err := utils.CalcFileMd5(optFile)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", md5str)
	return nil
}

// sumCmd represents the 'smc sum' command
var sumCmd = &cobra.Command{
	Use:   "sum",
	Short: "Calculate MD5 checksum",
	Long:  `Calculate MD5 checksum for a file`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := md5sum(); err != nil {
			fmt.Println(err)
		}
	},
}

var optFile string

func init() {
	certCmd.AddCommand(sumCmd)

	sumCmd.Example = `  # Calculate file checksum using MD5 algorithm
  smc cert sum -f ./shenma`
	sumCmd.Flags().SortFlags = false
	sumCmd.Flags().StringVarP(&optFile, "file", "f", "", "File name")
	sumCmd.MarkFlagRequired("file")
}
