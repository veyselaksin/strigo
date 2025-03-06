package ratelimiter

import (
	"context"
	"fmt"

	"github.com/veyselaksin/strigo"
	"github.com/veyselaksin/strigo/internal/db"
)

// RateLimiter is the core struct that handles rate limiting logic
// It combines storage, configuration and strategy to implement rate limiting
type RateLimiter struct {
	storage  db.Storage     // Interface for storing rate limit data (Redis, Memcached, etc.)
	config   *strigo.Config // Configuration for rate limiting rules
	strategy Strategy       // Strategy interface for different rate limiting algorithms
}

// New creates a new rate limiter instance with the provided storage and configuration
// It validates the config and initializes the appropriate rate limiting strategy
func New(storage db.Storage, cfg *strigo.Config) (*RateLimiter, error) {
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
	// Get current count
	count, err := rl.storage.Get(ctx, key)
	if err != nil {
		return false
	}

	// Check if allowed using the strategy
	if !rl.strategy.IsAllowed(count, rl.config.Limit) {
		return false
	}

	// If allowed, increment the counter
	_, err = rl.storage.Increment(ctx, key, rl.config.GetDuration())
	return err == nil
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
