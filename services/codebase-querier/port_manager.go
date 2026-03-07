package main

import (
	"errors"
	"sync"
)

// PortManager 管理端口分配和配额
type PortManager struct {
	mu        sync.Mutex
	allocated map[int]bool
	maxConns  int // 最大并发数
	current   int // 当前并发数
}

// NewPortManager 创建新的端口管理器
func NewPortManager(maxConns int) *PortManager {
	return &PortManager{
		allocated: make(map[int]bool),
		maxConns:  maxConns,
	}
}

// AllocatePort 分配可用端口（含配额检查）
func (pm *PortManager) AllocatePort() (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 配额检查
	if pm.current >= pm.maxConns {
		return 0, errors.New("超过最大并发连接数")
	}

	// 端口分配逻辑
	for port := 8000; port < 9000; port++ {
		if !pm.allocated[port] {
			pm.allocated[port] = true
			pm.current++
			return port, nil
		}
	}
	return 0, errors.New("无可用端口")
}

// ReleasePort 释放端口并更新配额
func (pm *PortManager) ReleasePort(port int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.allocated[port] {
		delete(pm.allocated, port)
		pm.current--
	}
}
