package component

import (
	"costrict-keeper/internal/env"
	"costrict-keeper/internal/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var optRemoveComponent string

var removeCmd = &cobra.Command{
	Use:   "remove {component | -n component}",
	Short: "Remove the specified package",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Determine component name: prioritize positional arguments, then use command line arguments
		component := optRemoveComponent
		if len(args) > 0 && args[0] != "" {
			component = args[0]
		}

		if component == "" {
			fmt.Println("Error: Component name must be specified")
			return
		}

		if err := removeComponent(component); err != nil {
			fmt.Println(err)
		}
	},
}

/**
 * Remove specified component package
 * @param {string} component - Name of the component to remove
 * @returns {error} Returns error if removal fails, nil on success
 * @description
 * - Uses RemovePackage function to remove component files and metadata
 * - Handles both configuration and executable package types
 * - Provides user feedback on removal status
 * @throws
 * - Package file removal errors
 * - Package description file removal errors
 */
func removeComponent(component string) error {
	// Call RemovePackage function to remove package
	u := utils.NewUpgrader(component, utils.UpgradeConfig{
		BaseDir: env.CostrictDir,
	}, nil)
	defer u.Close()

	if err := u.RemovePackage(nil); err != nil {
		fmt.Printf("Failed to remove component '%s': %v\n", component, err)
		return err
	}

	fmt.Printf("Component '%s' has been successfully removed\n", component)
	return nil
}

func init() {
	removeCmd.Flags().SortFlags = false
	removeCmd.Flags().StringVarP(&optRemoveComponent, "component", "n", "", "Specify the component name to remove")
	componentCmd.AddCommand(removeCmd)
}
