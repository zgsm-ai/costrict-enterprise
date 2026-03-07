package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/zgsm-ai/codebase-indexer/internal/tracer"
	"time"

	redsync "github.com/go-redsync/redsync/v4"
	redsyngoredis "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredis "github.com/redis/go-redis/v9"
)

// DistributedLock 分布式锁接口 - 修改为返回 Mutex 实例
type DistributedLock interface {
	// TryLock 尝试获取指定键的锁，并设置过期时间。
	// 如果成功获取锁，返回 mutex 实例和 true，否则返回 nil 和 false。如果发生错误，则返回错误。
	TryLock(ctx context.Context, key string, expiration time.Duration) (*redsync.Mutex, bool, error)
	// Lock 尝试获取指定键的锁，如果未立即获取到，则会阻塞直到获取到锁或 context 被取消。
	// 如果成功获取锁，返回 mutex 实例和 nil。如果获取锁失败或 context 被取消，则返回 nil 和错误。
	Lock(ctx context.Context, key string, expiration time.Duration) (*redsync.Mutex, error)
	// IsLocked 检查指定键的锁当前是否被持有。
	IsLocked(ctx context.Context, key string) (bool, error)
	// Unlock 释放指定键的锁。
	// 必须传入加锁时返回的 mutex 实例才能成功解锁。
	Unlock(ctx context.Context, mutex *redsync.Mutex) error
}

// redisDistLock 是基于 Redsync 的分布式锁管理器
type redisDistLock struct {
	rs     *redsync.Redsync
	client *goredis.Client
}

// NewRedisDistributedLock 创建一个新的 Redsync 分布式锁管理器实例。
func NewRedisDistributedLock(redisClient *goredis.Client) (DistributedLock, error) {
	// 使用 go-redis 客户端创建一个 Redsync 连接池
	pool := redsyngoredis.NewPool(redisClient)

	// 创建 Redsync 客户端
	rs := redsync.New(pool)

	// 返回包装了 Redsync 客户端的分布式锁管理器
	return &redisDistLock{
		rs:     rs,
		client: redisClient,
	}, nil
}

// TryLock 实现 DistributedLock 接口的 TryLock 方法
func (m *redisDistLock) TryLock(ctx context.Context, key string, expiration time.Duration) (*redsync.Mutex, bool, error) {
	// 为指定的 key 创建一个 Redsync mutex 实例
	mutex := m.rs.NewMutex(key, redsync.WithExpiry(expiration))

	// 尝试获取锁
	err := mutex.TryLockContext(ctx)

	// 根据 Redsync 的错误类型判断结果
	if err == nil {
		// 成功获取锁，返回 mutex 实例
		return mutex, true, nil
	} else if errors.Is(err, redsync.ErrFailed) {
		// 锁已被持有，尝试获取失败
		return nil, false, nil
	} else {
		// 发生其他错误
		return nil, false, fmt.Errorf("acquire lock failed, key: %s, err: %w", key, err)
	}
}

// Lock 实现 DistributedLock 接口的 Lock 方法
func (m *redisDistLock) Lock(ctx context.Context, key string, expiration time.Duration) (*redsync.Mutex, error) {
	// 为指定的 key 创建一个 Redsync mutex 实例
	mutex := m.rs.NewMutex(key, redsync.WithExpiry(expiration))

	// 尝试获取锁，会阻塞直到获取到或 context 被取消
	err := mutex.LockContext(ctx)
	if err != nil {
		// 获取锁失败
		return nil, fmt.Errorf("acquire lock failed, key: %s, err: %w", key, err)
	}
	// 成功获取锁，返回 mutex 实例
	return mutex, nil
}

// IsLocked 实现 DistributedLock 接口的 IsLocked 方法
func (m *redisDistLock) IsLocked(ctx context.Context, key string) (bool, error) {
	// 为指定的 key 创建一个 Redsync mutex 实例，使用一个非常短的过期时间
	mutex := m.rs.NewMutex(key, redsync.WithExpiry(1*time.Millisecond))

	// 尝试使用非阻塞方式获取锁
	err := mutex.TryLockContext(ctx)

	if err == nil {
		// 成功获取锁（说明之前未锁定），立即释放
		defer func() {
			if _, unlockErr := mutex.UnlockContext(ctx); unlockErr != nil {
				// 记录解锁临时锁时发生的错误
			}
		}()
		return false, nil // 未锁定
	} else if errors.Is(err, redsync.ErrFailed) {
		return true, nil // 已锁定
	} else {
		// 发生其他错误
		return false, fmt.Errorf("check lock failed, key: %s, err: %w", key, err)
	}
}

// Unlock 实现 DistributedLock 接口的 Unlock 方法
// 修改为接收 mutex 实例而不是 key
func (m *redisDistLock) Unlock(ctx context.Context, mutex *redsync.Mutex) error {
	if mutex == nil {
		return fmt.Errorf("mutex is nil")
	}

	// 释放锁。Redsync 会检查当前实例是否持有该锁。
	unlocked, err := mutex.UnlockContext(ctx)
	if errors.Is(err, redsync.ErrLockAlreadyExpired) {
		// 强行释放 TODO
		tracer.WithTrace(ctx).Debugf("redis_lock unlock failed with ErrLockAlreadyExpired, delete lock key %s force.", mutex.Name())
		if err = m.client.Del(ctx, mutex.Name()).Err(); err != nil {
			tracer.WithTrace(ctx).Errorf("redis_lock unlock force unlock failed, lock key %s.", mutex.Name())
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("release lock failed: %w", err)
	}

	// 如果 unlocked 为 false，同样表示锁不被当前实例持有或已释放
	if !unlocked {
		return fmt.Errorf("current node not own the lock or lock has been unlocked")
	}

	return nil // 成功释放锁
}
