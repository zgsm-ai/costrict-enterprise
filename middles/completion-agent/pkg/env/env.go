package env

import (
	"os"
	"path/filepath"
)

/**
 * Get costrict directory path
 * @returns {string} Returns costrict directory path
 * @description
 * - 获取用户主目录下的.costrict目录路径
 * - 在Windows系统下为%USERPROFILE%/.costrict
 * - 在Linux/macOS系统下为$HOME/.costrict
 * - 用于存储应用配置文件和日志
 */
func GetCostrictDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".costrict")
}

var DebugMode bool = false
