package strigo

import (
	"fmt"
	"time"
)

// Strategy represents the rate limiting strategy type
type Strategy string

// Available rate limiting strategies
const (
	TokenBucket   Strategy = "token_bucket"   // Classic token bucket algorithm
	LeakyBucket   Strategy = "leaky_bucket"   // Leaky bucket algorithm  
	FixedWindow   Strategy = "fixed_window"   // Fixed time window counting
	SlidingWindow Strategy = "sliding_window" // Sliding time window counting
)

// Options represents the rate limiter configuration options
// Inspired by rate-limiter-flexible package design
type Options struct {
	// Points defines the maximum number of points that can be consumed over duration
	// Default: 5
	Points int64 `json:"points"`
	
	// Duration defines the time window for point consumption in seconds
	// Default: 1 (per second)
	Duration int64 `json:"duration"`
	
	// Strategy defines the rate limiting algorithm to use
	// Default: TokenBucket
	Strategy Strategy `json:"strategy,omitempty"`
	
	// BlockDuration defines how long to block key after limit exceeded (in seconds)
	// Default: same as Duration
	BlockDuration int64 `json:"blockDuration,omitempty"`
	
	// KeyPrefix is used to create unique keys in the storage backend
	// Default: "rl" (rate limiter)
	KeyPrefix string `json:"keyPrefix,omitempty"`
	
	// StoreClient is the Redis/Memcached client instance
	// If nil, uses in-memory storage
	StoreClient interface{} `json:"-"`
	
	// StoreType specifies the type of store client ("redis", "memcached", "memory")
	// Auto-detected if StoreClient is provided
	StoreType string `json:"storeType,omitempty"`
}

// NewOptions creates default options similar to rate-limiter-flexible
func NewOptions() *Options {
	return &Options{
		Points:        5,
		Duration:      1,
		Strategy:      TokenBucket,
		BlockDuration: 0, // Will be set to Duration if 0
		KeyPrefix:     "rl",
		StoreType:     "memory",
	}
}

// Validate validates the options and sets defaults
func (o *Options) Validate() error {
	if o.Points <= 0 {
		return fmt.Errorf("points must be positive, got %d", o.Points)
	}
	
	if o.Duration <= 0 {
		return fmt.Errorf("duration must be positive, got %d", o.Duration)
	}
	
	// Set block duration to duration if not specified
	if o.BlockDuration <= 0 {
		o.BlockDuration = o.Duration
	}
	
	// Set default key prefix
	if o.KeyPrefix == "" {
		o.KeyPrefix = "rl"
	}
	
	// Set default strategy
	if o.Strategy == "" {
		o.Strategy = TokenBucket
	}
	
	// Validate strategy
	switch o.Strategy {
	case TokenBucket, LeakyBucket, FixedWindow, SlidingWindow:
		// Valid strategies
	default:
		return fmt.Errorf("invalid strategy: %s", o.Strategy)
	}
	
	return nil
}

// GetDuration returns the duration as time.Duration
func (o *Options) GetDuration() time.Duration {
	return time.Duration(o.Duration) * time.Second
}

// GetBlockDuration returns the block duration as time.Duration  
func (o *Options) GetBlockDuration() time.Duration {
	return time.Duration(o.BlockDuration) * time.Second
} 