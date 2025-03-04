package ratelimiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/veyselaksin/strigo/pkg/config"
)

// Strategy defines the interface for rate limiting strategies
// Different algorithms can be implemented by satisfying this interface
type Strategy interface {
	// IsAllowed determines if a request should be allowed based on:
	// count: current number of requests
	// limit: maximum allowed requests for the period
	IsAllowed(count int64, limit int64) bool
}

// TokenBucketStrategy implements the simple token bucket algorithm
// This is the default strategy that allows requests until the limit is reached
type TokenBucketStrategy struct{}

// IsAllowed implements the token bucket algorithm
// Returns true if current count is within the limit
func (s *TokenBucketStrategy) IsAllowed(count int64, limit int64) bool {
	return count <= limit
}

// LeakyBucketStrategy implements the leaky bucket algorithm
type LeakyBucketStrategy struct {
	mu       sync.Mutex
	leakRate time.Duration
	lastLeak time.Time
	current  int64 // Track current tokens
}

func NewLeakyBucketStrategy(rate time.Duration) *LeakyBucketStrategy {
	return &LeakyBucketStrategy{
		leakRate: rate,
		lastLeak: time.Now(),
	}
}

func (s *LeakyBucketStrategy) IsAllowed(count int64, limit int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(s.lastLeak)

	// Calculate leaked tokens
	leaked := int64(elapsed / s.leakRate)

	// Update current count considering leaks
	s.current = max(0, count-leaked)
	s.lastLeak = now

	// Check if new request can be accommodated
	if s.current < limit {
		s.current++
		return true
	}
	return false
}

// FixedWindowStrategy implements the fixed window algorithm
type FixedWindowStrategy struct {
	mu          sync.Mutex
	windowStart time.Time
	windowSize  time.Duration
	current     int64
}

func NewFixedWindowStrategy(windowSize time.Duration) *FixedWindowStrategy {
	return &FixedWindowStrategy{
		windowStart: time.Now(),
		windowSize:  windowSize,
		current:     0,
	}
}

func (s *FixedWindowStrategy) IsAllowed(count int64, limit int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if now.Sub(s.windowStart) >= s.windowSize {
		// Reset window
		s.windowStart = now
		s.current = 1 // Reset and count this request
		return true
	}

	// Check if within limit for current window
	if s.current < limit {
		s.current++
		return true
	}
	return false
}

// SlidingWindowStrategy implements the sliding window algorithm
type SlidingWindowStrategy struct {
	mu         sync.Mutex
	windowSize time.Duration
	buckets    map[int64]int64 // timestamp -> count
}

func NewSlidingWindowStrategy(windowSize time.Duration) *SlidingWindowStrategy {
	return &SlidingWindowStrategy{
		windowSize: windowSize,
		buckets:    make(map[int64]int64),
	}
}

func (s *SlidingWindowStrategy) IsAllowed(count int64, limit int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	currentBucket := now.Unix()
	windowStart := now.Add(-s.windowSize)

	// Clean old buckets
	for timestamp := range s.buckets {
		if time.Unix(timestamp, 0).Before(windowStart) {
			delete(s.buckets, timestamp)
		}
	}

	// Calculate current window count
	var windowCount int64
	for _, count := range s.buckets {
		windowCount += count
	}

	// Check if new request can be accommodated
	if windowCount < limit {
		s.buckets[currentBucket]++
		return true
	}
	return false
}

// NewStrategy creates the appropriate strategy based on configuration
// Currently only implements TokenBucket, but can be extended for other algorithms
func NewStrategy(strategyType config.Strategy) (Strategy, error) {
	switch strategyType {
	case config.TokenBucket:
		return &TokenBucketStrategy{}, nil
	case config.LeakyBucket:
		return NewLeakyBucketStrategy(time.Second), nil // 1 token/second leak rate
	case config.FixedWindow:
		return NewFixedWindowStrategy(time.Minute), nil // 1-minute window
	case config.SlidingWindow:
		return NewSlidingWindowStrategy(time.Minute), nil // 1-minute sliding window
	default:
		return nil, fmt.Errorf("unsupported strategy: %s", strategyType)
	}
}

// Helper function for Go versions < 1.21
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
