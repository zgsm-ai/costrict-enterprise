package cmd

import (
	"fmt"

	"costrict-keeper/cmd/root"
	"costrict-keeper/internal/env"

	"github.com/spf13/cobra"
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

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  `The 'version' command shows version details including git commit and build time`,

	Run: func(cmd *cobra.Command, args []string) {
		PrintVersions()
	},
}

func init() {
	root.RootCmd.AddCommand(versionCmd)

	versionCmd.Example = `  costrict version`
	env.Version = SoftwareVer
}
