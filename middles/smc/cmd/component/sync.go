package component

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/internal/env"
	"github.com/zgsm-ai/smc/internal/utils"
)

// syncCmd represents the 'smc component sync' command
var syncCmd = &cobra.Command{
	Use:   "sync {package | -p package} --target target",
	Short: "Sync remote package directory to target directory",
	Long:  `Syncs the remote package directory structure to a local target directory`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optSyncPackageName = args[0]
		}
		if optSyncPackageName == "" {
			return fmt.Errorf("package name is required")
		}
		return syncPackage(optSyncPackageName, optSyncTarget)
	},
}

const syncExample = `  # Sync aip package to ./aip directory
  smc component sync aip --target ./aip
  # Sync aip package with flag
  smc component sync -p aip --target ./sync-output`

var optSyncPackageName string
var optSyncTarget string

func init() {
	componentCmd.AddCommand(syncCmd)
	syncCmd.Flags().SortFlags = false
	syncCmd.Example = syncExample
	syncCmd.Flags().StringVarP(&optSyncPackageName, "package", "p", "", "Package name")
	syncCmd.Flags().StringVar(&optSyncTarget, "target", "", "Target directory to sync to")
}

/**
 *	同步包到目标目录
 */
func syncPackage(packageName, targetDir string) error {
	u := utils.NewUpgrader(packageName, utils.UpgradeConfig{
		BaseUrl: env.BaseUrl + "/costrict",
	}, nil)
	defer u.Close()

	if targetDir == "" {
		targetDir = filepath.Join(u.BaseDir, "components", packageName)
	}
	err := u.SyncPackage(targetDir)
	if err != nil {
		return fmt.Errorf("sync package '%s' to '%s' failed: %v", packageName, targetDir, err)
	}
	fmt.Printf("Successfully synced package '%s' to '%s'\n", packageName, targetDir)
	return nil
}
