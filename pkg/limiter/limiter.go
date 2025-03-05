package limiter

import (
	"fmt"
	"strings"

	"context"

	"github.com/veyselaksin/strigo/internal/db"
	"github.com/veyselaksin/strigo/internal/ratelimiter"
	"github.com/veyselaksin/strigo/pkg/config"
	"github.com/veyselaksin/strigo/pkg/duration"
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
	Period   duration.Period
	Limit    int64
	Strategy config.Strategy
}

// Config holds the rate limiter configuration
type Config struct {
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
func NewLimiter(cfg Config) (Limiter, error) {
	// Create storage backend
	var storage db.Storage
	switch cfg.Backend {
	case Redis:
		redisClient, err := db.NewRedisClient(cfg.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to create Redis client: %w", err)
		}
		storage = redisClient
	case Memcached:
		memcachedClient, err := db.NewMemcachedClient(cfg.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to create Memcached client: %w", err)
		}
		storage = memcachedClient
	default:
		return nil, fmt.Errorf("unsupported backend: %s", cfg.Backend)
	}

	// Initialize rules map
	rules := make(map[string]RuleConfig)
	for _, rule := range cfg.Rules {
		rules[rule.Pattern] = rule
	}

	// Ensure default rule exists
	if cfg.Default.Period == "" {
		cfg.Default = RuleConfig{
			Period:   duration.MINUTELY,
			Limit:    100,
			Strategy: config.TokenBucket,
		}
	}

	return &limiterImpl{
		storage:  storage,
		rules:    rules,
		default_: cfg.Default,
		prefix:   cfg.Prefix,
	}, nil
}

// Allow checks if a request should be allowed
func (l *limiterImpl) Allow(key string) bool {
	// Find matching rule
	rule := l.findMatchingRule(key)

	// Create internal config for this request
	internalCfg := &config.Config{
		Strategy: rule.Strategy,
		Period:   rule.Period,
		Limit:    rule.Limit,
		Prefix:   l.prefix,
	}

	// Create rate limiter for this request
	rl, err := ratelimiter.New(l.storage, internalCfg)
	if err != nil {
		return false
	}

	return rl.Allow(key)
}

// findMatchingRule returns the matching rule for a given key
func (l *limiterImpl) findMatchingRule(key string) RuleConfig {
	for pattern, rule := range l.rules {
		if matchPattern(key, pattern) {
			return rule
		}
	}
	return l.default_
}

// matchPattern checks if a key matches a pattern
func matchPattern(key, pattern string) bool {
	return strings.HasPrefix(key, pattern)
}

// Reset resets the rate limit for a given key
func (l *limiterImpl) Reset(key string) error {
	return l.storage.Reset(context.Background(), key)
}

// Close closes the rate limiter
func (l *limiterImpl) Close() error {
	return l.storage.Close()
}
