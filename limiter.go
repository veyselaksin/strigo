package strigo

import (
	"fmt"
	"strings"
	"time"

	"context"

	"github.com/veyselaksin/strigo/internal/db"
)

// Backend represents the storage backend type
type Backend string

const (
	Redis     Backend = "redis"
	Memcached Backend = "memcached"
)

// RuleConfig represents a rate limit rule for a specific pattern
type RuleConfig struct {
	Pattern  string // Pattern to match against keys
	Period   Period
	Limit    int64
	Strategy Strategy
}

// Config holds the rate limiter configuration
type LimiterConfig struct {
	Backend Backend
	Address string
	Rules   []RuleConfig // Add rules for different patterns
	Default RuleConfig   // Default rule if no pattern matches
	Prefix  string
}

// Limiter interface defines the rate limiting operations
type Limiter interface {
	Allow(key string) bool
	Reset(key string) error
	Close() error
}

type limiterImpl struct {
	storage  db.Storage
	rules    map[string]RuleConfig
	default_ RuleConfig
	prefix   string
}

// NewLimiter creates a new rate limiter instance
func NewLimiter(cfg LimiterConfig) (Limiter, error) {
	var storage db.Storage
	var err error

	// Initialize storage backend
	switch cfg.Backend {
	case Redis:
		storage, err = db.NewRedisClient(cfg.Address)
	case Memcached:
		storage, err = db.NewMemcachedClient(cfg.Address)
	default:
		return nil, fmt.Errorf("unsupported backend: %s", cfg.Backend)
	}

	if err != nil {
		return nil, err
	}

	// Convert rules to map for faster lookup
	rules := make(map[string]RuleConfig)
	for _, rule := range cfg.Rules {
		rules[rule.Pattern] = rule
	}

	return &limiterImpl{
		storage:  storage,
		rules:    rules,
		default_: cfg.Default,
		prefix:   cfg.Prefix,
	}, nil
}

// Allow checks if the request should be allowed
func (l *limiterImpl) Allow(key string) bool {
	// Use default prefix if no prefix is set
	if l.prefix == "" {
		l.prefix = "strigo"
	}

	// Build Redis key with prefix
	redisKey := fmt.Sprintf("%s:%s", l.prefix, key)

	// Find matching rule
	var rule RuleConfig
	var matched bool
	for pattern, r := range l.rules {
		// Check if key contains the pattern
		if strings.Contains(key, pattern) {
			rule = r
			matched = true
			break
		}
	}

	// Use default rule if no pattern matches
	if !matched {
		if l.default_.Pattern != "" {
			rule = l.default_
		} else {
			// If no default rule and no match, allow the request
			return true
		}
	}

	// Get current period window
	window := time.Now().Truncate(rule.Period.ToDuration())
	redisKey = fmt.Sprintf("%s:%d", redisKey, window.Unix())

	// Get current count
	count, err := l.storage.Get(context.Background(), redisKey)
	if err != nil {
		return false
	}

	// Check if allowed
	if count < rule.Limit {
		// Increment counter
		_, err := l.storage.Increment(context.Background(), redisKey, rule.Period.ToDuration())
		if err != nil {
			return false
		}
		return true
	}

	return false
}

// Reset resets the rate limit for the given key
func (l *limiterImpl) Reset(key string) error {
	redisKey := fmt.Sprintf("%s:%s", l.prefix, key)
	return l.storage.Reset(context.Background(), redisKey)
}

// Close closes the rate limiter
func (l *limiterImpl) Close() error {
	return l.storage.Close()
}

// GetUniqueKey generates a unique key for the config
func (c *LimiterConfig) GetUniqueKey() string {
	// Create a unique key based on rules
	key := c.Prefix
	for _, rule := range c.Rules {
		key += ":" + rule.Pattern
	}
	return key
}
