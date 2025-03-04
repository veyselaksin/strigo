package ratelimiter

import "errors"

var (
	// ErrInvalidConfig is returned when the configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrStorageUnavailable is returned when the storage backend is unavailable
	ErrStorageUnavailable = errors.New("storage backend unavailable")

	// ErrRateLimitExceeded is returned when the rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)
