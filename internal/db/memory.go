package db

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// MemoryStorage provides an in-memory implementation of the Storage interface
// Useful for testing or when no external storage backend is available
type MemoryStorage struct {
	data   map[string]int64
	jsonData map[string][]byte
	expiry map[string]time.Time
	mu     sync.RWMutex
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		data:     make(map[string]int64),
		jsonData: make(map[string][]byte),
		expiry:   make(map[string]time.Time),
	}
	
	// Start cleanup goroutine
	go storage.cleanup()
	
	return storage
}

// Increment increments the counter for the given key by the specified amount and returns the new count
func (m *MemoryStorage) Increment(ctx context.Context, key string, amount int64, expiry time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if key has expired
	if exp, exists := m.expiry[key]; exists && time.Now().After(exp) {
		delete(m.data, key)
		delete(m.expiry, key)
	}
	
	// Increment counter by the specified amount
	count := m.data[key] + amount
	m.data[key] = count
	m.expiry[key] = time.Now().Add(expiry)
	
	return count, nil
}

// Get returns the current count for the given key
func (m *MemoryStorage) Get(ctx context.Context, key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Check if key has expired
	if exp, exists := m.expiry[key]; exists && time.Now().After(exp) {
		return 0, nil
	}
	
	return m.data[key], nil
}

// Reset resets the counter for the given key
func (m *MemoryStorage) Reset(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.data, key)
	delete(m.jsonData, key)
	delete(m.expiry, key)
	
	return nil
}

// SetJSON stores a JSON-serializable object with expiry
func (m *MemoryStorage) SetJSON(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	m.jsonData[key] = data
	m.expiry[key] = time.Now().Add(expiry)
	
	return nil
}

// GetJSON retrieves and deserializes a JSON object
func (m *MemoryStorage) GetJSON(ctx context.Context, key string, dest interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Check if key has expired
	if exp, exists := m.expiry[key]; exists && time.Now().After(exp) {
		return nil // Key expired, return empty
	}
	
	data, exists := m.jsonData[key]
	if !exists {
		return nil // Key doesn't exist, return empty
	}
	
	return json.Unmarshal(data, dest)
}

// Close closes the storage (no-op for memory storage)
func (m *MemoryStorage) Close() error {
	return nil
}

// cleanup removes expired keys periodically
func (m *MemoryStorage) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for key, exp := range m.expiry {
			if now.After(exp) {
				delete(m.data, key)
				delete(m.jsonData, key)
				delete(m.expiry, key)
			}
		}
		m.mu.Unlock()
	}
} 