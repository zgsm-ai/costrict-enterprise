//go:build darwin

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// SetNewPG 设置进程属性，使子进程在父进程退出后继续运行
// Darwin系统实现
func SetNewPG(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
}

// KillProcessByPID 根据PID杀死进程
func KillProcessByPID(pid int) error {
	// 查找进程
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process with PID %d: %v", pid, err)
	}

	// 在Darwin上，我们需要杀死整个进程组
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to process with PID %d: %v", pid, err)
	}

	return nil
}

// IsProcessRunning 检查进程是否正在运行
func IsProcessRunning(pid int) (bool, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, fmt.Errorf("failed to find process with PID %d: %v", pid, err)
	}

	// 在Darwin上，发送signal 0来检查进程是否存在
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// 如果进程不存在，os.SyscallError 会包含 "no such process" 或类似的错误信息
		return false, fmt.Errorf("process with PID %d is not running: %v", pid, err)
	}

	return true, nil
}

// 根据PID获取进程名
func GetProcessName(pid int) (string, error) {
	// 在Darwin系统上，使用ps命令获取进程名
	// 使用command字段替代comm字段，避免命令名被截断
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "command=")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get process name for PID %d: %v", pid, err)
	}

	// 去除头尾空白字符
	commandLine := strings.TrimSpace(string(output))
	if commandLine == "" {
		return "", fmt.Errorf("no process found with PID %d", pid)
	}

	// 从完整命令行中提取命令名称（第一段内容）
	fields := strings.Fields(commandLine)
	if len(fields) == 0 {
		return "", fmt.Errorf("invalid command format for PID %d", pid)
	}

	// 获取命令名称并去除路径
	processName := filepath.Base(fields[0])
	return processName, nil
}
