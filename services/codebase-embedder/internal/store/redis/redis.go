package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/zgsm-ai/codebase-indexer/internal/config"
)

// NewRedisClient creates a new Redis client.
func NewRedisClient(c config.RedisConfig) (*redis.Client, error) {
	// 构建原生Redis客户端配置
	rdbCfg := redis.Options{
		Addr:         c.Addr,
		Password:     c.Password,
		DB:           c.DB,
		PoolSize:     c.PoolSize,
		MinIdleConns: c.MinIdleConn,
		DialTimeout:  c.ConnectTimeout,
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
	}

	client := redis.NewClient(&rdbCfg)

	// 测试连接
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}
