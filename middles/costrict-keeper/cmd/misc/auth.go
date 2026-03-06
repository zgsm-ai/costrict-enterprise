package client

import (
	"fmt"

	"costrict-keeper/cmd/root"
	"costrict-keeper/internal/config"
	"costrict-keeper/internal/utils"

	"github.com/golang-jwt/jwt/v5"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Display costrict authentication configuration",
	Long:  `Display costrict authentication configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		showAuthConfigs()
	},
}

const authExample = `  # Show all authentication configs
  costrict auth`

func showAuthConfigs() {
	auth := config.GetAuthConfig()

	fmt.Printf("Base URL: %s\n", auth.BaseUrl)
	fmt.Printf("User ID: %s\n", auth.ID)
	fmt.Printf("User Name: %s\n", auth.Name)
	fmt.Printf("Machine ID: %s\n", auth.MachineID)
	fmt.Printf("Access Token: %s\n", auth.AccessToken)
	// Parse token without verification (for now)
	if optViewJwt {
		token, _, err := jwt.NewParser().ParseUnverified(auth.AccessToken, jwt.MapClaims{})
		if err == nil {
			fmt.Printf("============= JWT ==============\n")
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				utils.PrintYaml(claims)
			}
		}
	} else {
		fmt.Printf("Decoded JWT: run `costrict auth --jwt`\n")
	}
}

var optViewJwt bool

func init() {
	authCmd.Flags().SortFlags = false
	authCmd.Flags().BoolVarP(&optViewJwt, "jwt", "j", false, "Display the decoded JWT")
	authCmd.Example = authExample
	root.RootCmd.AddCommand(authCmd)
}
