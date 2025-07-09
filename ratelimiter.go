package strigo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/veyselaksin/strigo/v2/internal/db"
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
	
	// Dispatch to strategy-specific implementation
	switch rl.opts.Strategy {
	case TokenBucket:
		return rl.consumeTokenBucket(ctx, key, consumePoints)
	case LeakyBucket:
		return rl.consumeLeakyBucket(ctx, key, consumePoints)
	case SlidingWindow:
		return rl.consumeSlidingWindow(ctx, key, consumePoints)
	case FixedWindow:
		return rl.consumeFixedWindow(ctx, key, consumePoints)
	default:
		// Default to TokenBucket for unknown strategies
		return rl.consumeTokenBucket(ctx, key, consumePoints)
	}
}

// Get returns the current rate limit information for the given key without consuming points
// Similar to rateLimiter.get(key) from rate-limiter-flexible
func (rl *RateLimiter) Get(key string) (*Result, error) {
	ctx := context.Background()
	storageKey := rl.buildKey(key)
	
	// Strategy-specific get implementations
	switch rl.opts.Strategy {
	case TokenBucket:
		return rl.getTokenBucket(ctx, storageKey)
	case LeakyBucket:
		return rl.getLeakyBucket(ctx, storageKey)
	case SlidingWindow:
		return rl.getSlidingWindow(ctx, storageKey)
	case FixedWindow:
		return rl.getFixedWindow(ctx, storageKey)
	default:
		return rl.getTokenBucket(ctx, storageKey)
	}
}

// Strategy-specific Get implementations

func (rl *RateLimiter) getTokenBucket(ctx context.Context, storageKey string) (*Result, error) {
	dataKey := fmt.Sprintf("%s:tb", storageKey)
	var data TokenBucketData
	err := rl.storage.GetJSON(ctx, dataKey, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to get token bucket data: %w", err)
	}
	
	if data.LastRefill.IsZero() {
		return nil, nil // No data exists
	}
	
	// Calculate current tokens
	now := time.Now()
	elapsed := now.Sub(data.LastRefill).Seconds()
	tokensToAdd := elapsed * data.RefillRate
	currentTokens := data.Tokens + tokensToAdd
	if currentTokens > float64(data.Capacity) {
		currentTokens = float64(data.Capacity)
	}
	
	return &Result{
		MsBeforeNext:      0,
		RemainingPoints:   int64(currentTokens),
		ConsumedPoints:    data.Capacity - int64(currentTokens),
		IsFirstInDuration: false,
		TotalHits:         rl.opts.Points,
		Allowed:           int64(currentTokens) >= 1,
	}, nil
}

func (rl *RateLimiter) getLeakyBucket(ctx context.Context, storageKey string) (*Result, error) {
	dataKey := fmt.Sprintf("%s:lb", storageKey)
	var data LeakyBucketData
	err := rl.storage.GetJSON(ctx, dataKey, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaky bucket data: %w", err)
	}
	
	if data.LastDrain.IsZero() {
		return nil, nil // No data exists
	}
	
	// Calculate current queue size after drainage
	now := time.Now()
	elapsed := now.Sub(data.LastDrain).Seconds()
	requestsToDrain := int64(elapsed * data.DrainRate)
	currentQueue := rl.drainRequests(data.Queue, requestsToDrain)
	
	currentPoints := int64(0)
	for _, req := range currentQueue {
		currentPoints += req.Points
	}
	
	return &Result{
		MsBeforeNext:      0,
		RemainingPoints:   rl.opts.Points - currentPoints,
		ConsumedPoints:    currentPoints,
		IsFirstInDuration: false,
		TotalHits:         rl.opts.Points,
		Allowed:           currentPoints < rl.opts.Points,
	}, nil
}

func (rl *RateLimiter) getSlidingWindow(ctx context.Context, storageKey string) (*Result, error) {
	dataKey := fmt.Sprintf("%s:sw", storageKey)
	var data SlidingWindowData
	err := rl.storage.GetJSON(ctx, dataKey, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to get sliding window data: %w", err)
	}
	
	if data.Requests == nil || len(data.Requests) == 0 {
		return nil, nil // No data exists
	}
	
	// Remove old requests outside window
	now := time.Now()
	windowStart := now.Add(-rl.opts.GetDuration())
	validRequests := rl.removeOldRequests(data.Requests, windowStart)
	
	return &Result{
		MsBeforeNext:      0,
		RemainingPoints:   rl.opts.Points - int64(len(validRequests)),
		ConsumedPoints:    int64(len(validRequests)),
		IsFirstInDuration: false,
		TotalHits:         rl.opts.Points,
		Allowed:           int64(len(validRequests)) < rl.opts.Points,
	}, nil
}

func (rl *RateLimiter) getFixedWindow(ctx context.Context, storageKey string) (*Result, error) {
	// Get current window information
	windowStart := rl.getWindowStartFixed()
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
	
	// Reset all strategy-specific keys
	strategies := []string{"tb", "lb", "sw"}
	for _, strategy := range strategies {
		dataKey := fmt.Sprintf("%s:%s", storageKey, strategy)
		_ = rl.storage.Reset(ctx, dataKey) // Ignore errors for non-existent keys
	}
	
	// Also reset the base key (for fixed window and backward compatibility)
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
	
	// Set a very high count to block all requests
	blockAmount := rl.opts.Points + 1000
	_, err := rl.storage.Increment(ctx, blockKey, blockAmount, duration)
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

// Deprecated: getWindowStart is replaced by strategy-specific implementations
// This method is kept for backward compatibility but should not be used
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