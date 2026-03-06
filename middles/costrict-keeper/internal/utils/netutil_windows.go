//go:build windows

package utils

import (
	"context"
	"fmt"
	"net"
	"syscall"
)

// checks if a port is listenable on localhost (Windows implementation)
func CheckPortListenable(port int) bool {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}

	// Create ListenConfig with control function to disable SO_REUSEADDR
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				// Disable SO_REUSEADDR to prevent address reuse
				syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 0)
			})
		},
	}

	l, err := lc.Listen(context.Background(), "tcp", addr.String())
	if err != nil {
		return false
	}
	defer l.Close()
	return true
}
