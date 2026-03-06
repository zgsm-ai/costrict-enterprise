package server

import (
	"costrict-keeper/internal/logger"
	"net"
	"os"
	"path/filepath"
	"runtime"
)

type ListenAddr struct {
	Network string
	Address string
}

/**
 * Test if the system supports Unix socket network type
 * @returns {bool} Returns true if Unix socket is supported, false otherwise
 * @description
 * - Creates a temporary Unix socket to test system support
 * - Cleans up test socket file after testing
 * - Returns false if Unix socket creation fails
 * - Returns true if Unix socket creation succeeds
 * @example
 * supported := IsUnixSocketSupported()
 * if !supported {
 *     logger.Info("Unix socket is not supported on this system")
 * }
 */
func IsUnixSocketSupported() bool {
	if runtime.GOOS != "windows" { //window,linux,darwin
		return true
	}
	// 尝试创建一个临时的Unix socket来测试系统是否支持
	testSocketPath := filepath.Join(os.TempDir(), "test_unix_socket.sock")
	// 清理可能存在的测试socket文件
	os.Remove(testSocketPath)

	// 尝试创建Unix socket监听器
	listener, err := net.Listen("unix", testSocketPath)
	if err != nil {
		// 如果创建失败，说明系统不支持Unix socket
		return false
	}

	// 如果创建成功，关闭监听器并清理文件
	listener.Close()
	os.Remove(testSocketPath)
	return true
}

/**
 * Create TCP and Unix socket listeners for cross-platform support
 * @param {[]ListenAddr} addrs - Listener Address
 * @returns {[]net.Listener} Array of created listeners
 * @returns {string} Unix socket path if created
 * @returns {error} Error if listener creation fails
 * @description
 * - Creates TCP listener if TCPPort > 0
 * - Creates Unix socket listener if SocketName is not empty
 * - Automatically determines platform-specific socket directory
 * - Cleans up existing socket files before creating new ones
 * - Sets appropriate file permissions for Unix socket
 * - Supports Windows, Linux, and Darwin platforms
 * @throws
 * - TCP listener creation errors
 * - Unix socket listener creation errors
 * - Socket file cleanup errors
 */
func CreateListeners(addrs []ListenAddr) ([]net.Listener, error) {
	var listeners []net.Listener

	var lastErr error
	for _, addr := range addrs {
		if addr.Network == "unix" {
			if err := os.Remove(addr.Address); err != nil && !os.IsNotExist(err) {
				logger.Errorf("Failed to remove existing socket file: %v", err)
				continue
			}
		}
		tcpListener, err := net.Listen(addr.Network, addr.Address)
		if err != nil {
			logger.Errorf("Failed to create listener on %s://%s: %v", addr.Network, addr.Address, err)
			lastErr = err
			continue
		}
		listeners = append(listeners, tcpListener)
	}
	return listeners, lastErr
}

// {
// 	// 创建Unix socket监听器
// 	if cfg.SocketName != "" {
// 		// 确定socket目录
// 		socketDir := cfg.SocketDir
// 		if socketDir == "" {
// 			socketDir = filepath.Join(env.CostrictDir, "run")
// 		}

// 		socketPath = filepath.Join(socketDir, cfg.SocketName)

// 		// 删除已存在的socket文件
// 		if _, err := os.Stat(socketPath); err == nil {
// 			os.Remove(socketPath)
// 		}

// 		unixListener, err := net.Listen("unix", socketPath)
// 		if err != nil {
// 			// 关闭已创建的TCP监听器
// 			for _, listener := range listeners {
// 				listener.Close()
// 			}
// 			return nil, "", fmt.Errorf("failed to create Unix socket listener on %s: %w", socketPath, err)
// 		}

// 		// 设置socket文件权限
// 		os.Chmod(socketPath, 0777)

// 		listeners = append(listeners, unixListener)
// 	}

// 	return listeners, socketPath, nil
// }
