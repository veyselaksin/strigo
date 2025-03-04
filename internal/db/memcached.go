package db

import (
	"context"
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type MemcachedClient struct {
	client *memcache.Client
}

func NewMemcachedClient(address string) (*MemcachedClient, error) {
	client := memcache.New(address)
	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to memcached: %w", err)
	}

	return &MemcachedClient{
		client: client,
	}, nil
}

func (m *MemcachedClient) Increment(ctx context.Context, key string, expiry time.Duration) (int64, error) {
	// Memcached increment returns uint64, we need to cast safely
	value, err := m.client.Increment(key, 1)
	if err == memcache.ErrCacheMiss {
		// Key doesn't exist, create it
		err = m.client.Set(&memcache.Item{
			Key:        key,
			Value:      []byte("1"),
			Expiration: int32(expiry.Seconds()),
		})
		if err != nil {
			return 0, err
		}
		return 1, nil
	}
	return int64(value), err
}

func (m *MemcachedClient) Get(ctx context.Context, key string) (int64, error) {
	item, err := m.client.Get(key)
	if err == memcache.ErrCacheMiss {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	// Parse the value as int64
	var value int64
	_, err = fmt.Sscanf(string(item.Value), "%d", &value)
	return value, err
}

func (m *MemcachedClient) Reset(ctx context.Context, key string) error {
	return m.client.Delete(key)
}

func (m *MemcachedClient) Close() error {
	// Memcache client doesn't have a close method
	return nil
}
