//go:build windows

package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// Windows API 常量和类型定义
const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
	PROCESS_TERMINATE         = 0x0001
	STILL_ACTIVE              = 259 // 进程仍在运行的标志
)

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	psapi                  = syscall.NewLazyDLL("psapi.dll")
	procOpenProcess        = kernel32.NewProc("OpenProcess")
	procCloseHandle        = kernel32.NewProc("CloseHandle")
	procEnumProcessModules = psapi.NewProc("EnumProcessModules")
	procGetModuleBaseNameW = psapi.NewProc("GetModuleBaseNameW")
	procTerminateProcess   = kernel32.NewProc("TerminateProcess")
	procGetExitCodeProcess = kernel32.NewProc("GetExitCodeProcess")
)

// SetNewPG 设置进程属性，使子进程在父进程退出后继续运行
// Windows系统实现
func SetNewPG(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// KillProcessByPID 根据PID杀死进程
func KillProcessByPID(pid int) error {
	// 打开进程句柄
	handle, _, err := procOpenProcess.Call(
		uintptr(PROCESS_TERMINATE),
		uintptr(0),
		uintptr(pid),
	)

	if handle == 0 {
		return fmt.Errorf("failed to open process with PID %d: %v", pid, err)
	}
	defer procCloseHandle.Call(handle)

	// 杀死进程
	ret, _, err := procTerminateProcess.Call(
		handle,
		uintptr(1),
	)

	if ret == 0 {
		return fmt.Errorf("failed to terminate process with PID %d: %v", pid, err)
	}

	return nil
}

// 根据PID获取进程名
func GetProcessName(pid int) (string, error) {
	// 打开进程句柄
	handle, _, _ := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ),
		uintptr(0),
		uintptr(pid),
	)

	if handle == 0 {
		return "", fmt.Errorf("failed to open process")
	}
	defer procCloseHandle.Call(handle)

	// 获取进程名
	var nameBuffer [260]uint16 // MAX_PATH
	var hModule uintptr

	// 首先枚举进程模块
	var cbNeeded uint32
	ret, _, err := procEnumProcessModules.Call(
		handle,
		uintptr(unsafe.Pointer(&hModule)),
		uintptr(unsafe.Sizeof(hModule)),
		uintptr(unsafe.Pointer(&cbNeeded)),
	)

	if ret == 0 {
		return "", fmt.Errorf("failed to enumerate modules: %v", err)
	}

	// 然后获取模块基础名称
	ret, _, err = procGetModuleBaseNameW.Call(
		handle,
		hModule,
		uintptr(unsafe.Pointer(&nameBuffer[0])),
		uintptr(len(nameBuffer)),
	)

	if ret == 0 {
		return "", fmt.Errorf("failed to get module base name: %v", err)
	}

	// 成功获取进程名
	processName := syscall.UTF16ToString(nameBuffer[:])
	return processName, nil
}

// IsProcessRunning 检查进程是否正在运行 使用 GetExitCodeProcess 检查进程是否正在运行
func IsProcessRunning(pid int) (bool, error) {
	// 打开进程句柄
	handle, _, err := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION),
		uintptr(0),
		uintptr(pid),
	)

	if handle == 0 {
		// 如果无法打开进程句柄，通常表示进程不存在
		return false, fmt.Errorf("failed to open process with PID %d: %v", pid, err)
	}
	defer procCloseHandle.Call(handle)

	// 获取进程退出码
	var exitCode uint32
	ret, _, err := procGetExitCodeProcess.Call(
		handle,
		uintptr(unsafe.Pointer(&exitCode)),
	)

	if ret == 0 {
		return false, fmt.Errorf("failed to get exit code for process with PID %d: %v", pid, err)
	}

	// 如果退出码是 STILL_ACTIVE，则进程仍在运行
	return exitCode == STILL_ACTIVE, nil
}

/**
 * Kill processes on Windows system
 * @param {string} processName - Name of the process to kill
 * @returns {error} Returns error if process killing fails, nil on success
 * @description
 * - Uses tasklist command to enumerate processes
 * - Parses output to find target process PIDs
 * - Kills each found process using taskkill command
 * @throws
 * - Command execution errors
 * - Process kill errors
 */
func KillSpecifiedProcess(processName string) error {
	log.Printf("Looking for process: %s\n", processName)
	// 获取所有进程ID和对应的进程名
	// 由于Windows API限制，我们需要使用其他方法来枚举进程
	// 这里使用tasklist命令作为备用方案
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s.exe", processName), "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to list processes for %s: %v\n", processName, err)
		return err
	}

	// 解析输出，获取PID
	lines := strings.Split(string(output), "\n")
	selfPid := os.Getpid()

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// CSV格式: "进程名","PID","会话名","会话#","内存使用"
		fields := strings.Split(line, "\",\"")
		if len(fields) < 2 {
			continue
		}
		// 移除引号
		procName := strings.Trim(fields[0], "\"")
		pidStr := strings.Trim(fields[1], "\"")
		procName = Path2ProcessName(procName)
		// 重复校验一下，检查进程名是否匹配，防止出错
		if !strings.EqualFold(procName, processName) {
			continue
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}
		if pid == selfPid {
			continue
		}

		// 使用Windows API杀死进程
		if err := KillProcessByPID(pid); err != nil {
			log.Printf("Failed to kill process %s (PID: %d): %v\n", processName, pid, err)
		} else {
			log.Printf("Successfully killed process %s (PID: %d)\n", processName, pid)
		}
	}
	return nil
}

func FindProcesses(processName string) []int {
	var pids []int
	// 获取所有进程ID和对应的进程名
	// 由于Windows API限制，我们需要使用其他方法来枚举进程
	// 这里使用tasklist命令作为备用方案
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s.exe", processName), "/FO", "CSV", "/NH")
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

		// CSV格式: "进程名","PID","会话名","会话#","内存使用"
		fields := strings.Split(line, "\",\"")
		if len(fields) < 2 {
			continue
		}
		// 移除引号
		procName := strings.Trim(fields[0], "\"")
		pidStr := strings.Trim(fields[1], "\"")
		procName = Path2ProcessName(procName)
		// 重复校验一下，检查进程名是否匹配，防止出错
		if !strings.EqualFold(procName, processName) {
			continue
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids
}
