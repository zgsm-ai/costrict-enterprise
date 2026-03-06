package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ------------------------------------------------------------------------------
//
//	进程名processName：
//		指进程的程序文件名，去掉路径和后缀并转成全小写后的名字
//		该名字已经去掉了各平台特有的“装饰”，是平台通用的名字
//		进程名一般可以在进程列表中获取到，但需要先进行一定处理
//
// ------------------------------------------------------------------------------
func Path2ProcessName(processPath string) string {
	base := filepath.Base(processPath)
	ext := filepath.Ext(processPath)
	return strings.ToLower(base[:len(base)-len(ext)])
}

/**
 *	根据进程名和PID查找并打开进程
 */
func FindProcess(processName string, pid int) (*os.Process, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}

	// 获取进程名
	name, err := GetProcessName(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to get process name for PID %d: %v", pid, err)
	}

	// 比较进程名（不区分大小写）
	if strings.EqualFold(name, processName) {
		return proc, nil
	}
	return nil, fmt.Errorf("process name mismatch: expected '%s', got '%s'", processName, name)
}

// 根据进程名和PID杀死进程
func KillProcess(processName string, pid int) error {
	proc, err := FindProcess(processName, pid)
	if err != nil {
		return err
	}
	if err = proc.Kill(); err != nil {
		return err
	}
	return nil
}

/**
 * Kill processes with specified names: costrict, codebase-indexer, cotun
 * @returns {error} Returns error if process killing fails, nil on success
 * @description
 * - Defines target process names
 * - Finds and kills processes by name for Windows and Unix systems
 * - Logs success and failure for each process
 * @throws
 * - Process enumeration errors
 * - Process kill errors
 */
func KillSpecifiedProcesses(targetProcesses []string) error {
	var last error
	for _, processName := range targetProcesses {
		if err := KillSpecifiedProcess(processName); err != nil {
			last = err
		}
	}
	return last
}
