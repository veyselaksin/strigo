package limiter

import (
	"fmt"

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

// Config holds the rate limiter configuration
type Config struct {
	Backend  Backend
	Address  string
	Period   duration.Period
	Limit    int64
	Prefix   string
	Strategy config.Strategy
}

// Limiter interface defines the rate limiting operations
type Limiter interface {
	Allow(key string) bool
	Reset(key string) error
	Close() error
}

type limiterImpl struct {
	rateLimiter *ratelimiter.RateLimiter
}

// NewLimiter creates a new rate limiter instance
func NewLimiter(cfg Config) (Limiter, error) {

	// Create internal config
	internalCfg := &config.Config{
		Strategy: config.TokenBucket, // Default strategy
		Period:   cfg.Period,
		Limit:    cfg.Limit,
		Prefix:   "strigo",
		BackendConfig: config.BackendConfig{
			Type:    string(cfg.Backend),
			Address: cfg.Address,
		},
	}

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

	// Create rate limiter
	rl, err := ratelimiter.New(storage, internalCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limiter: %w", err)
	}

	return &limiterImpl{
		rateLimiter: rl,
	}, nil
}

// Allow checks if a request should be allowed
func (l *limiterImpl) Allow(key string) bool {
	return l.rateLimiter.Allow(key)
}

// Reset resets the rate limit for a given key
func (l *limiterImpl) Reset(key string) error {
	return l.rateLimiter.Reset(key)
}

// Close closes the rate limiter
func (l *limiterImpl) Close() error {
	return l.rateLimiter.Close()
}
