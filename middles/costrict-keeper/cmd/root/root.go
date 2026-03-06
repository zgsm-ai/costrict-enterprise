package root

import (
	"costrict-keeper/internal/env"
	"fmt"

	"github.com/spf13/cobra"
)

var costrictPath string

var RootCmd = &cobra.Command{
	Use:   "costrict",
	Short: "Mobile CLI application manager",
	Long:  `costrict manages download, installation, startup, configuration, monitoring and service registration for multiple CLI programs`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if costrictPath != "" {
			env.CostrictDir = costrictPath
			fmt.Printf("Using a custom costrict directory: %s\n", costrictPath)
		}
	},
}

func init() {
	// Add global config option
	RootCmd.PersistentFlags().StringVarP(&costrictPath, "costrict", "c", "", "Specify the costrict data directory")
}
