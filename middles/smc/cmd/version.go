/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
)

var SoftwareVer = ""
var BuildTime = ""
var BuildTag = ""
var BuildCommitId = ""

func PrintVersions() {
	fmt.Printf("Version %s\n", SoftwareVer)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Build Tag: %s\n", BuildTag)
	fmt.Printf("Build Commit ID: %s\n", BuildCommitId)
}

// versionCmd represents the 'smc version' command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display smc version information",
	Long:  `The 'smc version' command shows version details including git commit and build time`,

	Run: func(cmd *cobra.Command, args []string) {
		PrintVersions()
	},
}

func init() {
	common.RootCmd.AddCommand(versionCmd)

	versionCmd.Example = `  smc version`
}
