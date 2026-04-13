package resilience

import (
	"context"
	"electric-backend/config"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// CacheGet retrieves a cached value from Redis. Returns nil if not found.
func CacheGet(key string, dest interface{}) bool {
	if config.RedisClient == nil {
		return false
	}
	val, err := config.RedisClient.Get(context.Background(), key).Result()
	if err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false
	}
	return true
}

// CacheSet stores a value in Redis with TTL.
func CacheSet(key string, value interface{}, ttl time.Duration) {
	if config.RedisClient == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	if err := config.RedisClient.Set(context.Background(), key, data, ttl).Err(); err != nil {
		log.Printf("⚠️ Cache set error [%s]: %v", key, err)
	}
}

// CacheInvalidate removes a cached key.
func CacheInvalidate(keys ...string) {
	if config.RedisClient == nil {
		return
	}
	config.RedisClient.Del(context.Background(), keys...)
}

// CacheKey builds a namespaced cache key.
func CacheKey(prefix string, parts ...string) string {
	key := "cache:" + prefix
	for _, p := range parts {
		key += ":" + p
	}
	return key
}

// CacheGetOrSet retrieves from cache, or calls fn and caches the result.
func CacheGetOrSet(key string, ttl time.Duration, dest interface{}, fn func() (interface{}, error)) error {
	if CacheGet(key, dest) {
		return nil
	}

	result, err := fn()
	if err != nil {
		return err
	}

	// Marshal result into dest
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("cache unmarshal error: %w", err)
	}

	CacheSet(key, result, ttl)
	return nil
}
