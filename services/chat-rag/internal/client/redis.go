package client

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zgsm-ai/chat-rag/internal/config"
)

// RedisInterface defines the interface for Redis client
type RedisInterface interface {
	// Connect establishes a connection to Redis
	Connect(ctx context.Context) error

	// SetHashField sets a field-value pair in a Redis hash
	SetHashField(ctx context.Context, key string, field string, value interface{}, expiration time.Duration) error

	// GetHashField retrieves a field value from a Redis hash
	GetHashField(ctx context.Context, key string, field string) (string, error)

	// GetHash retrieves all field-value pairs from a Redis hash
	GetHash(ctx context.Context, key string) (map[string]string, error)

	// HashLen returns the number of fields in a hash
	HashLen(ctx context.Context, key string) (int64, error)

	// GetString retrieves a string value by key
	GetString(ctx context.Context, key string) (string, error)

	// Close gracefully closes the Redis connection
	Close() error
}

// RedisClient handles communication with Redis
type RedisClient struct {
	client *redis.Client
	config config.RedisConfig
}

// NewRedisClient creates a new Redis client instance and connects to Redis
func NewRedisClient(cfg config.RedisConfig) RedisInterface {
	client := &RedisClient{
		config: cfg,
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		return client
	}

	return client
}

// Connect establishes a connection to Redis
func (c *RedisClient) Connect(ctx context.Context) error {
	c.client = redis.NewClient(&redis.Options{
		Addr:     c.config.Addr,
		Password: c.config.Password,
		DB:       c.config.DB,
	})

	// Ping to test connection
	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// SetHashField sets a field-value pair in a Redis hash
func (c *RedisClient) SetHashField(ctx context.Context, key string, field string, value interface{}, expiration time.Duration) error {
	if c.client == nil {
		if err := c.Connect(ctx); err != nil {
			return fmt.Errorf("redis client not connected and failed to reconnect: %w", err)
		}
	}

	err := c.client.HSet(ctx, key, field, value).Err()
	if err != nil {
		return fmt.Errorf("failed to set hash field in Redis: %w", err)
	}

	if expiration > 0 {
		err = c.client.Expire(ctx, key, expiration).Err()
		if err != nil {
			return fmt.Errorf("failed to set expiration for hash key: %w", err)
		}
	}

	return nil
}

// GetHashField retrieves a field value from a Redis hash
func (c *RedisClient) GetHashField(ctx context.Context, key string, field string) (string, error) {
	if c.client == nil {
		if err := c.Connect(ctx); err != nil {
			return "", fmt.Errorf("redis client not connected and failed to reconnect: %w", err)
		}
	}

	value, err := c.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("hash field does not exist: %s:%s", key, field)
		}
		return "", fmt.Errorf("failed to get hash field from Redis: %w", err)
	}

	return value, nil
}

// Close gracefully closes the Redis connection
func (c *RedisClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// GetHash retrieves all field-value pairs from a Redis hash
func (c *RedisClient) GetHash(ctx context.Context, key string) (map[string]string, error) {
	if c.client == nil {
		if err := c.Connect(ctx); err != nil {
			return nil, fmt.Errorf("redis client not connected and failed to reconnect: %w", err)
		}
	}

	values, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash from Redis: %w", err)
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("hash does not exist: %s", key)
	}

	return values, nil
}

// HashLen returns the number of fields in a hash
func (c *RedisClient) HashLen(ctx context.Context, key string) (int64, error) {
	if c.client == nil {
		if err := c.Connect(ctx); err != nil {
			return 0, fmt.Errorf("redis client not connected and failed to reconnect: %w", err)
		}
	}

	length, err := c.client.HLen(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get hash length from Redis: %w", err)
	}

	return length, nil
}

// GetString retrieves a string value by key
func (c *RedisClient) GetString(ctx context.Context, key string) (string, error) {
	if c.client == nil {
		if err := c.Connect(ctx); err != nil {
			return "", fmt.Errorf("redis client not connected and failed to reconnect: %w", err)
		}
	}

	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key does not exist: %s", key)
		}
		return "", fmt.Errorf("failed to get key from Redis: %w", err)
	}

	return value, nil
}
