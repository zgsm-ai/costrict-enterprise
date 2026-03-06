package env

import (
	"os"
	"path/filepath"
)

var Daemon bool = false
var ListenPort int = 0
var Version string = ""

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
