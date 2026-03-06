package rc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	"github.com/go-redis/redis/v8"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
)

/**
 * Initialize Redis client
 * @param {string} addr - Redis server address
 * @param {string} password - Redis password
 * @param {number} db - Redis database number
 * @returns {error} Error if connection fails
 */
func InitRedis(addr, password string, db int) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	if _, err := Client.Ping(Ctx).Result(); err != nil {
		return errors.Wrap(err, "failed to connect to redis")
	}
	return nil
}

/**
 * Set JSON data to Redis
 * @param {string} key - Redis key
 * @param {interface{}} value - Value to be stored as JSON
 * @param {Duration} expiration - Key expiration time
 * @returns {error} Error if operation fails
 */
func SetJSON(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "failed to marshal value")
	}
	return Client.Set(Ctx, key, data, expiration).Err()
}

/**
 * Get JSON data from Redis
 * @param {string} key - Redis key
 * @param {interface{}} dest - Destination object to store data
 * @returns {error} Error if operation fails or key not found
 */
func GetJSON(key string, dest interface{}) error {
	data, err := Client.Get(Ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return errors.Wrap(err, "failed to get value")
	}
	return json.Unmarshal(data, dest)
}

/**
 * Delete key from Redis
 * @param {string} key - Redis key to delete
 * @returns {error} Error if operation fails
 */
func Del(key string) error {
	return Client.Del(Ctx, key).Err()
}

/**
 * Check if key exists in Redis
 * @param {string} key - Redis key to check
 * @returns {boolean} Whether key exists
 * @returns {error} Error if operation fails
 */
func Exists(key string) (bool, error) {
	n, err := Client.Exists(Ctx, key).Result()
	return n > 0, err
}

/**
 * Get all matching keys by prefix pattern
 * @param {string} prefix - Key prefix pattern
 * @returns {[]string} List of matching keys
 * @returns {error} Error if scan operation fails
 */
func KeysByPrefix(prefix string) ([]string, error) {
	var keys []string
	var cursor uint64
	var err error

	for {
		// Safely iterate keys using SCAN command
		var partialKeys []string
		partialKeys, cursor, err = Client.Scan(Ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan keys")
		}

		keys = append(keys, partialKeys...)

		if cursor == 0 { // Iteration completed
			break
		}
	}

	return keys, nil
}

/**
 * Load all key-value pairs under specified prefix in Redis
 * @param {string} prefix - Key prefix pattern
 * @returns {Object} Map of key-value pairs
 * @returns {error} Error if operation fails
 */
func LoadJsons(prefix string) (map[string]interface{}, error) {
	keys, err := KeysByPrefix(prefix)
	if err != nil {
		return nil, err
	}

	jsons := make(map[string]interface{})
	for _, key := range keys {
		var val interface{}
		if err := GetJSON(key, &val); err != nil {
			return nil, err
		}
		jsons[key] = val
	}

	return jsons, nil
}
