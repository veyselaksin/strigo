package strigo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/veyselaksin/strigo/internal/db"
)

// RateLimiter provides rate limiting functionality similar to rate-limiter-flexible
type RateLimiter struct {
	storage db.Storage
	opts    *Options
}

// New creates a new rate limiter instance with the given options
// Similar to new RateLimiterMemory(opts) from rate-limiter-flexible
func New(opts *Options) (*RateLimiter, error) {
	if opts == nil {
		opts = NewOptions()
	}
	
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}
	
	// Initialize storage backend
	storage, err := initStorage(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	
	return &RateLimiter{
		storage: storage,
		opts:    opts,
	}, nil
}

// Consume attempts to consume the specified points for the given key
// If no points are specified, defaults to 1 point
func (rl *RateLimiter) Consume(key string, points ...int64) (*Result, error) {
	// Default to 1 point if not specified
	consumePoints := int64(1)
	if len(points) > 0 {
		consumePoints = points[0]
	}

	// Validate points
	if consumePoints < 0 {
		return nil, fmt.Errorf("points cannot be negative")
	}

	ctx := context.Background()
	storageKey := rl.buildKey(key)
	
	// Get current window information
	windowStart := rl.getWindowStart()
	windowKey := fmt.Sprintf("%s:%d", storageKey, windowStart.Unix())
	
	// Get current count from storage
	currentCount, err := rl.storage.Get(ctx, windowKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get current count: %w", err)
	}
	
	// Check if this is the first request in the window
	isFirstInDuration := currentCount == 0
	
	// Calculate if the request should be allowed
	newCount := currentCount + consumePoints
	allowed := newCount <= rl.opts.Points
	
	// Calculate remaining points
	remainingPoints := rl.opts.Points - currentCount
	if remainingPoints < 0 {
		remainingPoints = 0
	}
	
	// Calculate time until next window
	nextWindow := windowStart.Add(rl.opts.GetDuration())
	msBeforeNext := time.Until(nextWindow).Milliseconds()
	
	// If allowed, increment the counter
	consumedPoints := currentCount
	if allowed {
		_, err = rl.storage.Increment(ctx, windowKey, rl.opts.GetDuration())
		if err != nil {
			return nil, fmt.Errorf("failed to increment counter: %w", err)
		}
		consumedPoints = newCount
		remainingPoints = rl.opts.Points - newCount
		if remainingPoints < 0 {
			remainingPoints = 0
		}
	}
	
	result := &Result{
		MsBeforeNext:      msBeforeNext,
		RemainingPoints:   remainingPoints,
		ConsumedPoints:    consumedPoints,
		IsFirstInDuration: isFirstInDuration,
		TotalHits:         rl.opts.Points,
		Allowed:           allowed,
	}
	
	return result, nil
}

// Get returns the current rate limit information for the given key without consuming points
// Similar to rateLimiter.get(key) from rate-limiter-flexible
func (rl *RateLimiter) Get(key string) (*Result, error) {
	ctx := context.Background()
	storageKey := rl.buildKey(key)
	
	// Get current window information
	windowStart := rl.getWindowStart()
	windowKey := fmt.Sprintf("%s:%d", storageKey, windowStart.Unix())
	
	// Get current count from storage
	currentCount, err := rl.storage.Get(ctx, windowKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get current count: %w", err)
	}
	
	// If no data exists, return nil (similar to rate-limiter-flexible)
	if currentCount == 0 {
		return nil, nil
	}
	
	// Calculate remaining points
	remainingPoints := rl.opts.Points - currentCount
	if remainingPoints < 0 {
		remainingPoints = 0
	}
	
	// Calculate time until next window
	nextWindow := windowStart.Add(rl.opts.GetDuration())
	msBeforeNext := time.Until(nextWindow).Milliseconds()
	
	result := &Result{
		MsBeforeNext:      msBeforeNext,
		RemainingPoints:   remainingPoints,
		ConsumedPoints:    currentCount,
		IsFirstInDuration: false,
		TotalHits:         rl.opts.Points,
		Allowed:           currentCount <= rl.opts.Points,
	}
	
	return result, nil
}

// Reset resets the rate limit for the given key
// Similar to rateLimiter.delete(key) from rate-limiter-flexible
func (rl *RateLimiter) Reset(key string) error {
	ctx := context.Background()
	storageKey := rl.buildKey(key)
	return rl.storage.Reset(ctx, storageKey)
}

// Block blocks the key for the specified duration in seconds
// Similar to rateLimiter.block(key, secDuration) from rate-limiter-flexible
func (rl *RateLimiter) Block(key string, durationSec int64) error {
	ctx := context.Background()
	storageKey := rl.buildKey(key)
	
	// Set a high count that will block requests
	blockKey := fmt.Sprintf("%s:block", storageKey)
	duration := time.Duration(durationSec) * time.Second
	
	_, err := rl.storage.Increment(ctx, blockKey, duration)
	return err
}

// Close closes the rate limiter and cleans up resources
func (rl *RateLimiter) Close() error {
	if rl.storage != nil {
		return rl.storage.Close()
	}
	return nil
}

// buildKey creates the full storage key with prefix
func (rl *RateLimiter) buildKey(key string) string {
	return fmt.Sprintf("%s:%s", rl.opts.KeyPrefix, key)
}

// getWindowStart returns the start time of the current window based on strategy
func (rl *RateLimiter) getWindowStart() time.Time {
	now := time.Now()
	duration := rl.opts.GetDuration()
	
	switch rl.opts.Strategy {
	case FixedWindow:
		// Truncate to window boundary
		return now.Truncate(duration)
	case SlidingWindow:
		// For sliding window, use current time
		return now
	default:
		// For token bucket and leaky bucket, use fixed window approach
		return now.Truncate(duration)
	}
}

// initStorage initializes the appropriate storage backend
func initStorage(opts *Options) (db.Storage, error) {
	// If no store client provided, use memory storage
	if opts.StoreClient == nil {
		return db.NewMemoryStorage(), nil
	}
	
	// Auto-detect client type or use explicit store type
	switch {
	case opts.StoreType == "redis" || isRedisClient(opts.StoreClient):
		return db.NewRedisStorageFromClient(opts.StoreClient)
	case opts.StoreType == "memcached" || isMemcachedClient(opts.StoreClient):
		return db.NewMemcachedStorageFromClient(opts.StoreClient)
	default:
		return db.NewMemoryStorage(), nil
	}
}

// Helper functions to detect client types
func isRedisClient(client interface{}) bool {
	clientType := fmt.Sprintf("%T", client)
	return strings.Contains(clientType, "redis")
}

func isMemcachedClient(client interface{}) bool {
	clientType := fmt.Sprintf("%T", client)
	return strings.Contains(clientType, "memcache")
} 