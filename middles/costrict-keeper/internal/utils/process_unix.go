//go:build unix || linux || darwin

package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/**
 * Kill processes on Unix system (Linux/macOS)
 * @param {string} processName - Name of the process to kill
 * @returns {error} Returns error if process killing fails, nil on success
 * @description
 * - Uses ps command to enumerate processes with compatible format for both Linux and Darwin
 * - Parses output to find target process PIDs
 * - Implements graceful termination: first SIGTERM, then SIGKILL if needed
 * - Handles permission issues for both root and normal users
 * @throws
 * - Command execution errors
 * - Process kill errors
 */
func KillSpecifiedProcess(processName string) error {
	log.Printf("Looking for process: %s\n", processName)

	selfPid := os.Getpid()

	// 使用兼容Linux和Darwin的ps命令格式
	// -e: 显示所有进程，-o: 自定义输出格式
	// 使用command字段替代comm字段，避免命令名被截断
	cmd := exec.Command("ps", "-e", "-o", "pid,command")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to list processes for %s: %v\n", processName, err)
		return err
	}

	// 解析输出，获取PID
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// 跳过标题行
		if strings.HasPrefix(line, "PID") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pidStr := fields[0]
		procName := Path2ProcessName(fields[1])
		// 检查进程名是否匹配（不区分大小写）
		if !strings.EqualFold(procName, processName) {
			continue
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Printf("Failed to parse PID %s for process %s: %v\n", pidStr, processName, err)
			continue
		}
		if pid == selfPid {
			continue
		}

		// 优雅地杀死进程
		if err := killProcessGracefully(pid, processName); err != nil {
			log.Printf("Failed to kill process %s (PID: %d): %v\n", processName, pid, err)
		} else {
			log.Printf("Successfully killed process %s (PID: %d)\n", processName, pid)
		}
	}
	return nil
}

/**
 * Kill process gracefully with SIGTERM first, then SIGKILL if needed
 * @param {int} pid - Process ID to kill
 * @param {string} procName - Process name for logging
 * @returns {error} Returns error if process killing fails, nil on success
 * @description
 * - First tries to terminate process with SIGTERM (graceful shutdown)
 * - If SIGTERM fails or times out, uses SIGKILL (forceful termination)
 * - Handles permission errors appropriately
 * @throws
 * - Process not found errors
 * - Permission denied errors
 * - Signal sending errors
 */
func killProcessGracefully(pid int, procName string) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %s (PID: %d): %v", procName, pid, err)
	}

	// 首先尝试优雅终止 (SIGTERM)
	log.Printf("Attempting graceful termination of process %s (PID: %d)\n", procName, pid)
	err = process.Signal(syscall.SIGTERM)
	if err == nil {
		// 等待进程退出
		for i := 0; i < 10; i++ {
			// 检查进程是否还在运行
			if err := process.Signal(syscall.Signal(0)); err != nil {
				// 进程已退出
				log.Printf("Process %s (PID: %d) terminated gracefully\n", procName, pid)
				return nil
			}
			// 等待100ms
			time.Sleep(100 * time.Millisecond)
		}
	}

	// 如果SIGTERM失败，使用强制终止 (SIGKILL)
	log.Printf("Graceful termination failed, force killing process %s (PID: %d)\n", procName, pid)
	err = process.Signal(syscall.SIGKILL)
	if err != nil {
		return fmt.Errorf("failed to kill process %s (PID: %d): %v", procName, pid, err)
	}

	log.Printf("Process %s (PID: %d) force killed\n", procName, pid)
	return nil
}

func FindProcesses(processName string) []int {
	var pids []int

	// 使用兼容Linux和Darwin的ps命令格式
	// -e: 显示所有进程，-o: 自定义输出格式
	// 使用command字段替代comm字段，避免命令名被截断
	cmd := exec.Command("ps", "-e", "-o", "pid,command")
	output, err := cmd.Output()
	if err != nil {
		return pids
	}

	// 解析输出，获取PID
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 跳过标题行
		if strings.HasPrefix(line, "PID") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pidStr := fields[0]
		procName := Path2ProcessName(fields[1])
		// 检查进程名是否匹配（不区分大小写）
		if !strings.EqualFold(procName, processName) {
			continue
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Printf("Failed to parse PID %s for process %s: %v\n", pidStr, procName, err)
			continue
		}
		pids = append(pids, pid)
	}
	return pids
}
