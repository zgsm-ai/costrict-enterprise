package client

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/zgsm-ai/smc/internal/env"
	"github.com/zgsm-ai/smc/internal/utils"

	"github.com/spf13/cobra"
)

// GetLoginParamsFromFile creates LoginParams from me.json file, falling back to defaults
func getDefaultLoginParams() *utils.LoginParams {
	params := &utils.LoginParams{
		MachineCode:   "default-a1b2c3d4e5f6789012345678",
		VSCodeVersion: "1.101.0",
		PluginVersion: "2.0.6",
		Provider:      "casdoor",
		URIScheme:     "vscode",
		BaseURL:       "https://zgsm.sangfor.com",
	}

	// Use values from config if they're not empty
	if env.BaseUrl != "" {
		params.BaseURL = env.BaseUrl
	}
	if env.MachineId != "" {
		params.MachineCode = env.MachineId
	}
	if env.PluginVersion != "" {
		params.PluginVersion = env.PluginVersion
	}
	if env.VscodeVersion != "" {
		params.VSCodeVersion = env.VscodeVersion
	}

	return params
}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate via OAuth and save access token",
	Long: `Complete the OAuth authentication process and save access token:
1. Opens the browser with the login URL
2. Periodically polls the token endpoint until a valid token is received
3. Decodes JWT token to get user information
4. Saves authentication configuration to local file

The URL will be constructed with the following parameters:
- machine_code: Machine identifier (from option or 'smc config list machineId')
- plugin_version: Plugin version (from option or 'smc config list pluginVersion')
- vscode_version: VSCode version (from option or 'smc config list vscodeVersion')
- base_url: Base URL for login endpoint (from option or 'smc config list baseUrl')

After successful authentication, the access token and user information will be saved to .costrict/share/auth.json file`,
	Run: func(cmd *cobra.Command, args []string) {
		machineCode, _ := cmd.Flags().GetString("machine-code")
		pluginVersion, _ := cmd.Flags().GetString("plugin-version")
		vscodeVersion, _ := cmd.Flags().GetString("vscode-version")
		baseURL, _ := cmd.Flags().GetString("base-url")

		// First try to load from me.json file
		params := getDefaultLoginParams()

		// Override with command line values if provided
		params.State = generateRandomState()

		if machineCode != "" {
			params.MachineCode = machineCode
		}
		if pluginVersion != "" {
			params.PluginVersion = pluginVersion
		}
		if vscodeVersion != "" {
			params.VSCodeVersion = vscodeVersion
		}
		if baseURL != "" {
			params.BaseURL = baseURL
		}
		// Perform complete login process
		token, err := utils.Login(params, func() error {
			fmt.Print(".")
			return nil
		})
		if err != nil {
			fmt.Printf("Login failed: %v\n", err)
			return
		}

		fmt.Printf("Access token: %s\n", token)

		// Decode JWT token to get user information
		claims, err := utils.DecodeJWT(token)
		if err != nil {
			fmt.Printf("Warning: Failed to decode JWT token: %v\n", err)
			return
		}
		fmt.Printf("%+v", claims)
		// Create auth configuration
		authConfig := utils.AuthConfig{
			ID:          claims.ID,
			Name:        claims.DisplayName,
			AccessToken: token,
			MachineID:   params.MachineCode,
			BaseUrl:     params.BaseURL,
		}

		// Save authentication configuration
		if err := utils.SaveAuthConfig(authConfig); err != nil {
			fmt.Printf("Warning: Failed to save auth config: %v\n", err)
			return
		}

		fmt.Printf("Authentication configuration saved to .costrict/share/auth.json\n")
	},
}

func init() {
	// Add login command to root command
	clientCmd.AddCommand(loginCmd)

	// Add flags for login command
	loginCmd.Flags().String("base-url", "", "Base URL for login endpoint (default: https://zgsm.sangfor.com)")
	loginCmd.Flags().String("machine-code", "", "Machine identifier (default: from me.json)")
	loginCmd.Flags().String("plugin-version", "", "Plugin version (default: from me.json)")
	loginCmd.Flags().String("vscode-version", "", "VSCode version (default: from me.json)")
}

// generateRandomState generates a 16-byte random string for OAuth state parameter
func generateRandomState() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simple default state if random generation fails
		return "default_state_" + fmt.Sprintf("%d", len(bytes))
	}
	return hex.EncodeToString(bytes)
}
