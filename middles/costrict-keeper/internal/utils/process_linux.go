//go:build linux

package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// SetNewPG 设置进程属性，使子进程在父进程退出后继续运行
// Linux系统实现
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

	// 在Linux上，我们需要杀死整个进程组
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

	// 在Linux上，发送signal 0来检查进程是否存在
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// 如果进程不存在，os.SyscallError 会包含 "no such process" 或类似的错误信息
		return false, fmt.Errorf("process with PID %d is not running: %v", pid, err)
	}

	return true, nil
}

// GetProcessName 根据PID获取进程名
func GetProcessName(pid int) (string, error) {
	// 读取/proc/<pid>/cmdline文件
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmdline, err := ioutil.ReadFile(cmdlinePath)
	if err != nil {
		return "", fmt.Errorf("failed to read cmdline for PID %d: %v", pid, err)
	}

	if len(cmdline) == 0 {
		return "", fmt.Errorf("no cmdline found for PID %d", pid)
	}

	// 分割参数（以null字符分隔）
	args := strings.Split(string(cmdline), "\x00")
	if len(args) == 0 {
		return "", fmt.Errorf("invalid cmdline format for PID %d", pid)
	}

	// 获取可执行文件名
	execPath := args[0]
	return filepath.Base(execPath), nil
}
