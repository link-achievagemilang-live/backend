package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRepository handles caching operations
type RedisRepository struct {
	client *redis.Client
}

// NewRedisRepository creates a new Redis repository
func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

// Set caches a URL mapping with TTL
func (r *RedisRepository) Set(ctx context.Context, shortCode, originalURL string, ttl time.Duration) error {
	key := fmt.Sprintf("url:%s", shortCode)
	err := r.client.Set(ctx, key, originalURL, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache URL: %w", err)
	}
	return nil
}

// Get retrieves a URL from cache
func (r *RedisRepository) Get(ctx context.Context, shortCode string) (string, error) {
	key := fmt.Sprintf("url:%s", shortCode)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("cache miss")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get from cache: %w", err)
	}
	return val, nil
}

// Delete removes a URL from cache
func (r *RedisRepository) Delete(ctx context.Context, shortCode string) error {
	key := fmt.Sprintf("url:%s", shortCode)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}
	return nil
}

// Exists checks if a key exists in cache
func (r *RedisRepository) Exists(ctx context.Context, shortCode string) (bool, error) {
	key := fmt.Sprintf("url:%s", shortCode)
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}
	return count > 0, nil
}
