package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

/**
 * Client authentication configuration
 * @property {string} id - Client unique identifier
 * @property {string} name - Client display name
 * @property {string} access_token - JWT access token for authentication
 * @property {string} machine_id - Machine unique identifier
 * @property {string} base_url - Base URL for API endpoints
 */
type AuthConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AccessToken string `json:"access_token"`
	MachineID   string `json:"machine_id"`
	BaseUrl     string `json:"base_url"`
}

var (
	authConfig *AuthConfig
	authLock   sync.RWMutex
)

/**
 * Load client configuration from auth.json file
 * @returns {error} Returns error if loading fails, nil on success
 * @description
 * - Loads client authentication configuration from .costrict/share/auth.json
 * - File contains client ID, name, access token, machine ID and base URL
 * - Configuration is cached in memory for subsequent calls
 */
func LoadAuthConfig() error {
	authPath := filepath.Join(getCostrictDir(), "share", "auth.json")

	if _, err := os.Stat(authPath); os.IsNotExist(err) {
		return fmt.Errorf("auth config file not found: %s", authPath)
	}

	file, err := os.Open(authPath)
	if err != nil {
		return fmt.Errorf("failed to open auth config file: %w", err)
	}
	defer file.Close()

	var newConfig AuthConfig
	if err := json.NewDecoder(file).Decode(&newConfig); err != nil {
		return fmt.Errorf("failed to decode auth config: %w", err)
	}

	authLock.Lock()
	defer authLock.Unlock()

	authConfig = &newConfig
	return nil
}

/**
 * Get client configuration instance
 * @returns {AuthConfig} Returns client configuration instance
 * @description
 * - Returns cached client configuration
 * - If configuration is not loaded, attempts to load it first
 * - Returns empty config if loading fails
 * @example
 * config := GetAuthConfig()
 * if config.ID == "" {
 *     log.Println("Client not configured")
 * }
 */
func GetAuthConfig() AuthConfig {
	authLock.RLock()
	if authConfig != nil {
		defer authLock.RUnlock()
		return *authConfig
	}
	authLock.RUnlock()

	if err := LoadAuthConfig(); err != nil {
		return AuthConfig{}
	}

	authLock.RLock()
	defer authLock.RUnlock()
	return *authConfig
}

/**
 * Check if client is configured
 * @returns {bool} Returns true if client is properly configured, false otherwise
 * @description
 * - Checks if client configuration has been loaded and contains required fields
 * - Required fields: ID, AccessToken, MachineID
 */
func IsAuthConfigured() bool {
	auth := GetAuthConfig()
	return auth.ID != "" && auth.AccessToken != "" && auth.MachineID != ""
}

func GetAuthHeader() (string, string) {
	return "Authorization", "Bearer " + GetAuthConfig().AccessToken
}

/**
 * Get base URL for API requests
 * @returns {string} Returns base URL or empty string if not configured
 * @description
 * - Returns the configured base URL for API endpoints
 * - Used to construct full API request URLs
 * @example
 * baseURL := GetBaseURL()
 * if baseURL != "" {
 *     apiURL := baseURL + "/api/v1/endpoint"
 * }
 */
func GetBaseURL() string {
	auth := GetAuthConfig()
	return auth.BaseUrl
}

/**
 * Get client machine ID
 * @returns {string} Returns machine ID or empty string if not configured
 * @description
 * - Returns the unique machine identifier from client configuration
 * - Used for machine-specific operations and authentication
 */
func GetMachineID() string {
	auth := GetAuthConfig()
	return auth.MachineID
}
