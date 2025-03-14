// Package strigo provides a comprehensive rate limiting implementation for Go applications,
// with support for multiple strategies and storage backends.
//
// # Overview
//
// StriGO is a flexible and efficient rate limiting library that supports multiple
// rate limiting strategies and storage backends. It is designed to be easy to use
// while providing powerful features for advanced use cases.
//
// Features
//
//   - Multiple rate limiting strategies
//   - Flexible storage backend support
//   - Pattern-based rate limiting rules
//   - Middleware support for popular web frameworks
//   - Thread-safe implementation
//   - Configurable time windows
//
// # Rate Limiting Strategies
//
// StriGO supports the following rate limiting strategies:
//
//   - TokenBucket: Classic token bucket algorithm with continuous rate limiting
//   - LeakyBucket: Leaky bucket algorithm for constant outflow rate
//   - FixedWindow: Simple counting in fixed time windows
//   - SlidingWindow: More accurate counting using sliding time windows
//
// # Storage Backends
//
// The package supports multiple storage backends:
//
//   - Redis: For distributed rate limiting
//   - Memcached: Alternative distributed storage
//
// Basic Usage
//
//	// Create a new rate limiter with default configuration
//	config := strigo.NewDefaultConfig()
//	limiter, err := strigo.NewRateLimiter(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Check if request is allowed
//	allowed, err := limiter.Allow("user-123")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Advanced Configuration
//
// For more complex rate limiting scenarios, you can use pattern-based rules:
//
//	config := &strigo.LimiterConfig{
//		Backend: strigo.Redis,
//		Address: "localhost:6379",
//		Rules: []strigo.RuleConfig{
//			{
//				Pattern:  "api",
//				Period:   strigo.MINUTELY,
//				Limit:    100,
//				Strategy: strigo.TokenBucket,
//			},
//			{
//				Pattern:  "admin",
//				Period:   strigo.HOURLY,
//				Limit:    1000,
//				Strategy: strigo.SlidingWindow,
//			},
//		},
//		Default: strigo.RuleConfig{
//			Period:   strigo.DAILY,
//			Limit:    10000,
//			Strategy: strigo.FixedWindow,
//		},
//		Prefix: "myapp",
//	}
//
// # Time Periods
//
// StriGO supports various time periods for rate limiting:
//
//   - SECONDLY: Per second rate limiting
//   - MINUTELY: Per minute rate limiting
//   - HOURLY: Per hour rate limiting
//   - DAILY: Per day rate limiting
//
// # Web Framework Integration
//
// StriGO provides middleware support for popular web frameworks:
//
//   - Fiber
//   - Standard net/http
//   - Echo (coming soon)
//   - Gin (coming soon)
//
// Example with Fiber:
//
//	app := fiber.New()
//	config := strigo.NewDefaultConfig()
//	limiter, err := strigo.NewRateLimiter(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	app.Use(middleware.NewFiberMiddleware(limiter))
//
// # Error Handling
//
// StriGO provides detailed error information for various scenarios:
//
//   - Configuration validation errors
//   - Backend connection errors
//   - Rate limit exceeded errors
//   - Storage operation errors
//
// # Thread Safety
//
// All StriGO operations are thread-safe and can be safely used in concurrent applications.
// The storage backends handle concurrent access appropriately.
//
// Performance Considerations
//
//   - Use Redis for high-throughput applications
//   - Consider using the TokenBucket strategy for smooth rate limiting
//   - Optimize pattern matching by using specific patterns
//   - Use appropriate time windows based on your use case
//
// For more detailed examples and documentation, visit:
// https://veyselaksin.github.io/strigo/
//
// For bug reports and feature requests, visit:
// https://github.com/veyselaksin/strigo/issues
package strigo
