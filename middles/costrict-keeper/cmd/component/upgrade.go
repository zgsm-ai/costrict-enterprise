package component

import (
	"costrict-keeper/internal/config"
	"costrict-keeper/internal/env"
	"costrict-keeper/internal/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var optComponent string
var optVersion string

var upgradeCmd = &cobra.Command{
	Use:   "upgrade {component | -n component}",
	Short: "Upgrade specified component",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Determine component name: prioritize positional arguments, then use command line arguments
		component := optComponent
		if len(args) > 0 && args[0] != "" {
			component = args[0]
		}

		if component == "" {
			fmt.Println("Error: Component name must be specified")
			return
		}

		upgradeComponent(component, optVersion)
	},
}

func upgradeComponent(component string, version string) error {
	u := utils.NewUpgrader(component, utils.UpgradeConfig{
		BaseUrl: config.Cloud().UpgradeUrl,
		BaseDir: env.CostrictDir,
	}, nil)
	defer u.Close()

	var specVer *utils.VersionNumber
	if version != "" {
		var v utils.VersionNumber
		if err := v.Parse(version); err != nil {
			fmt.Printf("Invalid version number: %s\n", version)
			return err
		}
		specVer = &v
	}

	pkg, upgraded, err := u.UpgradePackage(specVer)
	if err != nil {
		fmt.Printf("The '%s' upgrade failed: %v\n", component, err)
		return err
	}
	if !upgraded {
		fmt.Printf("The '%s' version is up to date\n", component)
	} else {
		fmt.Printf("The '%s' is upgraded to version %s\n", component, pkg.VersionId.String())
	}
	return nil
}

func init() {
	upgradeCmd.Flags().SortFlags = false
	upgradeCmd.Flags().StringVarP(&optVersion, "version", "v", "", "Specify the target version to upgrade")
	upgradeCmd.Flags().StringVarP(&optComponent, "component", "n", "", "Specify the component name to upgrade")
	componentCmd.AddCommand(upgradeCmd)
}
