package utils

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/zgsm-ai/smc/internal/env"
)

// TokenResponse represents the response from the token endpoint
type TokenResponse struct {
	Code int `json:"code,omitempty"`
	Data struct {
		AccessToken  string `json:"access_token,omitempty"`
		RefreshToken string `json:"refresh_token,omitempty"`
		State        string `json:"state,omitempty"`
	} `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Success bool   `json:"success,omitempty"`
}

// LoginParams represents the parameters for building login URL
type LoginParams struct {
	BaseURL       string `json:"base_url,omitempty"`
	MachineCode   string `json:"machine_code,omitempty"`
	PluginVersion string `json:"plugin_version,omitempty"`
	VSCodeVersion string `json:"vscode_version,omitempty"`
	State         string `json:"state,omitempty"`
	Provider      string `json:"provider,omitempty"`
	URIScheme     string `json:"uri_scheme,omitempty"`
}
type AuthConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AccessToken string `json:"access_token"`
	MachineID   string `json:"machine_id"`
	BaseUrl     string `json:"base_url"`
}

// (default: %USERPROFILE%/.costrict on Windows, $HOME/.costrict on Linux)
var CostrictDir string = GetCostrictDir()

/**
 * Get costrict directory path
 * @returns {string} Returns costrict directory path
 */
func GetCostrictDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".costrict")
}

// "/oidc-auth/api/v1/plugin/login/token"
// "/oidc-auth/api/v1/plugin/login"
// BuildLoginURL constructs the login URL with the given parameters
// @Summary Build login URL with optional parameter ignoring
// @Description Constructs a login URL using the provided parameters, optionally ignoring specified URL parameters
// @param params - Login parameters object containing all login configuration
// @param path - Path to append to the base URL (optional, defaults to "/login")
// @param ignores - String array of parameter names to exclude from the URL (optional)
// @returns Built login URL as string, or error if URL construction fails
func BuildLoginURL(params *LoginParams, path string, ignores ...string) (string, error) {
	if params == nil {
		return "", fmt.Errorf("params cannot be nil")
	}

	// Create a map of parameters to be ignored for quick lookup
	ignoreMap := make(map[string]bool)
	for _, ignore := range ignores {
		ignoreMap[ignore] = true
	}

	// Base URL - use the one from params or default
	baseURL := params.BaseURL
	if baseURL == "" {
		baseURL = "https://zgsm.sangfor.com"
	}

	// Create URL object
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}
	// Append path to the base URL
	u.Path = u.Path + path

	// Create query parameters
	q := u.Query()

	// Add parameters if they are not empty and not in the ignore list
	if params.MachineCode != "" && !ignoreMap["machine_code"] {
		q.Set("machine_code", params.MachineCode)
	}

	if params.State != "" && !ignoreMap["state"] {
		q.Set("state", params.State)
	}

	if params.Provider != "" && !ignoreMap["provider"] {
		q.Set("provider", params.Provider)
	}

	if params.PluginVersion != "" && !ignoreMap["plugin_version"] {
		q.Set("plugin_version", params.PluginVersion)
	}

	if params.VSCodeVersion != "" && !ignoreMap["vscode_version"] {
		q.Set("vscode_version", params.VSCodeVersion)
	}

	if params.URIScheme != "" && !ignoreMap["uri_scheme"] {
		q.Set("uri_scheme", params.URIScheme)
	}

	// Set query parameters
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// OpenBrowser opens the specified URL in the default browser
func OpenBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		// On Windows, use rundll32 to open URL, which handles special characters properly
		// This avoids issues with '&' and other special characters in URLs
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{url}
	}
	return startProcess(cmd, args...)
}

// startProcess starts a process with the given command and arguments
func startProcess(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	return cmd.Start()
}

// GetToken sends a GET request to the token endpoint with the same parameters used for login
// @Summary Get authentication token from token endpoint
// @Description Sends a GET request to the token endpoint with login parameters
// @param params - Login parameters object containing all login configuration
// @returns TokenResponse object containing authentication token or error, or error if request fails
func GetToken(params *LoginParams) (*TokenResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	// Use BuildLoginURL to construct the token URL with the same parameters
	tokenURL, err := BuildLoginURL(params, "/oidc-auth/api/v1/plugin/login/token")
	if err != nil {
		return nil, fmt.Errorf("failed to build token URL: %w", err)
	}

	// Create a transport with insecure TLS verification (similar to other utils)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: env.SkipSSL},
	}
	client := &http.Client{Transport: tr}

	// Create HTTP request
	req, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send token request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}
	// Parse the JSON response
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// Login performs the complete login process:
// 1. Opens the browser with the login URL
// 2. Periodically polls the token endpoint until a valid token is received
// 3. Returns the access token or an error
func Login(params *LoginParams, progress func() error) (string, error) {
	if params == nil {
		return "", fmt.Errorf("params cannot be nil")
	}

	// 1. Open the browser with the login URL
	loginURL, err := BuildLoginURL(params, "/oidc-auth/api/v1/plugin/login")
	if err != nil {
		return "", fmt.Errorf("failed to build login URL: %w", err)
	}

	fmt.Printf("Opening login URL in browser: %s\n", loginURL)
	if err := OpenBrowser(loginURL); err != nil {
		fmt.Printf("WARN: failed to open browser: %w", err)
	}

	// 2. Periodically poll the token endpoint
	fmt.Println("Waiting for authentication completion...")

	// Set up polling with a 3-second interval, timeout after 5 minutes
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("login timed out after 5 minutes")
		case <-ticker.C:
			// Check for token
			if progress != nil {
				if err := progress(); err != nil {
					return "", err
				}
			}
			tokenResp, err := GetToken(params)
			if err != nil {
				// Continue polling even if there's an error getting the token
				fmt.Printf("Error checking token status: %v, retrying...\n", err)
				continue
			}

			// Check if we have a valid token
			if tokenResp != nil && tokenResp.Data.AccessToken != "" {
				fmt.Println("Authentication successful!")
				return tokenResp.Data.AccessToken, nil
			}

			// If we got an error response but no token, continue waiting
			if tokenResp != nil && tokenResp.Message != "" {
				fmt.Printf("Authentication not yet completed: %s, waiting...\n", tokenResp.Message)
			}
		}
	}
}

func SaveAuthConfig(config AuthConfig) error {
	// Ensure the directory exists
	authDir := filepath.Join(CostrictDir, "share")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		return fmt.Errorf("failed to create auth directory: %w", err)
	}

	// Create the file
	authPath := filepath.Join(authDir, "auth.json")
	file, err := os.Create(authPath)
	if err != nil {
		return fmt.Errorf("failed to create auth config file: %w", err)
	}
	defer file.Close()

	// Encode and write the configuration
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode auth config: %w", err)
	}
	return nil
}
