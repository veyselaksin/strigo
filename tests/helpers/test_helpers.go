package helpers

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/redis/go-redis/v9"
)

func getRedisAddress() string {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "6379"
	}
	return fmt.Sprintf("%s:%s", host, port)
}

func getMemcachedAddress() string {
	host := os.Getenv("MEMCACHED_HOST")
	port := os.Getenv("MEMCACHED_PORT")
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "11211"
	}
	return fmt.Sprintf("%s:%s", host, port)
}

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: getRedisAddress(),
	})
}

func NewMemcachedClient() *memcache.Client {
	return memcache.New(getMemcachedAddress())
}

func CleanupRedis(t *testing.T, rdb *redis.Client) {
	err := rdb.FlushAll(context.Background()).Err()
	if err != nil {
		t.Fatalf("Failed to cleanup Redis: %v", err)
	}
}

func CleanupMemcached(t *testing.T, mc *memcache.Client) {
	err := mc.FlushAll()
	if err != nil {
		t.Fatalf("Failed to cleanup Memcached: %v", err)
	}
}
