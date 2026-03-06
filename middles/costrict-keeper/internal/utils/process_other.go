//go:build !windows && !linux && !darwin

package utils

import (
	"os/exec"
)

// SetNewPG 设置进程属性，使子进程在父进程退出后继续运行
// 默认实现，用于不支持的构建目标
func SetNewPG(cmd *exec.Cmd) {
	// 默认不做任何处理
}

// KillProcess 根据进程名和PID杀死进程
// 默认实现，用于不支持的构建目标
func KillProcessByPID(pid int) error {
	panic("KillProcessByPID not implemented for this platform")
}

// IsProcessRunning 检查进程是否正在运行
func IsProcessRunning(pid int) (bool, error) {
	panic("IsProcessRunning not implemented for this platform")
}

// GetProcessName 根据PID获取进程名
func GetProcessName(pid int) (string, error) {
	panic("GetProcessName not implemented for this platform")
}
