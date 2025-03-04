package ratelimiter

import (
	"context"
	"fmt"

	"github.com/veyselaksin/strigo/internal/db"
	"github.com/veyselaksin/strigo/pkg/config"
)

// RateLimiter is the core struct that handles rate limiting logic
// It combines storage, configuration and strategy to implement rate limiting
type RateLimiter struct {
	storage  db.Storage     // Interface for storing rate limit data (Redis, Memcached, etc.)
	config   *config.Config // Configuration for rate limiting rules
	strategy Strategy       // Strategy interface for different rate limiting algorithms
}

// New creates a new rate limiter instance with the provided storage and configuration
// It validates the config and initializes the appropriate rate limiting strategy
func New(storage db.Storage, cfg *config.Config) (*RateLimiter, error) {
	// Validate configuration before creating the rate limiter
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create the rate limiting strategy based on configuration
	strategy, err := NewStrategy(cfg.Strategy)
	if err != nil {
		return nil, err
	}

	// Return new RateLimiter instance with all components initialized
	return &RateLimiter{
		storage:  storage,
		config:   cfg,
		strategy: strategy,
	}, nil
}

// Allow is a convenience method that checks if a request should be allowed
// It uses a background context for the check
func (rl *RateLimiter) Allow(key string) bool {
	ctx := context.Background()
	return rl.AllowWithContext(ctx, key)
}

// AllowWithContext checks if a request should be allowed with the provided context
func (rl *RateLimiter) AllowWithContext(ctx context.Context, key string) bool {
	// Create a unique key by combining prefix and provided key
	key = fmt.Sprintf("%s:%s", rl.config.Prefix, key)

	// Get current count
	count, err := rl.storage.Get(ctx, key)
	if err != nil {
		// If there's an error getting the count, be conservative and deny
		return false
	}

	// Check if allowed using the strategy
	if allowed := rl.strategy.IsAllowed(count, rl.config.Limit); allowed {
		// If allowed, increment the counter
		_, err = rl.storage.Increment(ctx, key, rl.config.GetDuration())
		return err == nil
	}

	return false
}

// Reset resets the rate limit for a given key
func (rl *RateLimiter) Reset(key string) error {
	ctx := context.Background()
	key = fmt.Sprintf("%s:%s", rl.config.Prefix, key)
	return rl.storage.Reset(ctx, key)
}

// Close cleans up any resources
func (rl *RateLimiter) Close() error {
	return rl.storage.Close()
}
