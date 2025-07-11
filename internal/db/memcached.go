package db

import (
	"context"
	"encoding/json"
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

func (m *MemcachedClient) Increment(ctx context.Context, key string, amount int64, expiry time.Duration) (int64, error) {
	// Memcached increment accepts uint64, need to convert safely
	if amount < 0 {
		return 0, fmt.Errorf("memcached increment amount cannot be negative: %d", amount)
	}
	
	value, err := m.client.Increment(key, uint64(amount))
	if err == memcache.ErrCacheMiss {
		// Key doesn't exist, create it with the initial amount
		err = m.client.Set(&memcache.Item{
			Key:        key,
			Value:      []byte(fmt.Sprintf("%d", amount)),
			Expiration: int32(expiry.Seconds()),
		})
		if err != nil {
			return 0, err
		}
		return amount, nil
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

// SetJSON stores a JSON-serializable object with expiry
func (m *MemcachedClient) SetJSON(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	return m.client.Set(&memcache.Item{
		Key:        key,
		Value:      data,
		Expiration: int32(expiry.Seconds()),
	})
}

// GetJSON retrieves and deserializes a JSON object
func (m *MemcachedClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	item, err := m.client.Get(key)
	if err == memcache.ErrCacheMiss {
		return nil // Key doesn't exist, return empty
	}
	if err != nil {
		return err
	}
	
	return json.Unmarshal(item.Value, dest)
}

func (m *MemcachedClient) Close() error {
	// Memcache client doesn't have a close method
	return nil
}

// NewMemcachedStorageFromClient creates a Memcached storage instance from an existing Memcached client
func NewMemcachedStorageFromClient(client interface{}) (Storage, error) {
	memcachedClient, ok := client.(*memcache.Client)
	if !ok {
		return nil, fmt.Errorf("invalid client type: expected *memcache.Client, got %T", client)
	}
	
	return &MemcachedClient{
		client: memcachedClient,
	}, nil
}
