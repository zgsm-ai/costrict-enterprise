package mocks

import (
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

// NewMockRedis 创建一个新的 mock Redis 客户端
func NewMockRedis() (redis.Cmdable, redismock.ClientMock) {
	return redismock.NewClientMock()
}
