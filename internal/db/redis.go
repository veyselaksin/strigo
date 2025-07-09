package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(address string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr: address,
	})

	// Test the connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{
		client: client,
	}, nil
}

func (r *RedisClient) Increment(ctx context.Context, key string, amount int64, expiry time.Duration) (int64, error) {
	pipe := r.client.Pipeline()
	incr := pipe.IncrBy(ctx, key, amount)
	pipe.Expire(ctx, key, expiry)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

func (r *RedisClient) Get(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (r *RedisClient) Reset(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// SetJSON stores a JSON-serializable object with expiry
func (r *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	return r.client.Set(ctx, key, data, expiry).Err()
}

// GetJSON retrieves and deserializes a JSON object
func (r *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil // Key doesn't exist, return empty
	}
	if err != nil {
		return err
	}
	
	return json.Unmarshal([]byte(val), dest)
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

// NewRedisStorageFromClient creates a Redis storage instance from an existing Redis client
func NewRedisStorageFromClient(client interface{}) (Storage, error) {
	redisClient, ok := client.(*redis.Client)
	if !ok {
		return nil, fmt.Errorf("invalid client type: expected *redis.Client, got %T", client)
	}
	
	return &RedisClient{
		client: redisClient,
	}, nil
}
