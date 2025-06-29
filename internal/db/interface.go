package db

import (
	"context"
	"time"
)

// Storage defines the interface for rate limiter storage backends
type Storage interface {
	// Increment increments the counter for the given key by the specified amount and returns the new count
	Increment(ctx context.Context, key string, amount int64, expiry time.Duration) (int64, error)

	// Get returns the current count for the given key
	Get(ctx context.Context, key string) (int64, error)

	// Reset resets the counter for the given key
	Reset(ctx context.Context, key string) error

	// Close closes the storage connection
	Close() error
}
