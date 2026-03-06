/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package common

import (
	"os"
	"fmt"

	"github.com/spf13/cobra"
)

/**
 * Root command configuration with base flags and examples
 */
var RootCmd = &cobra.Command{
	Use:   "smc",
	Short: "Shenma client tool",
	Long: `smc is the Shenma client tool for managing backend services
Features include task management, tagging, and extension management`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}

const rootExample = `  smc task list
  smc template add
  smc pool add
`

/**
 * Initialize root command with examples and persistent flags
 */
func init() {
	RootCmd.Example = rootExample
	RootCmd.PersistentFlags().StringVarP(&OptLogFile, "logfile", "L", "", "Log file path (empty for stderr, +xx.log for both stderr and file)")
	RootCmd.PersistentFlags().StringVarP(&OptDebug, "debug", "D", "", "Debug level: Off,Err,Dbg")
}
