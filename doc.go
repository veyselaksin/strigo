/*
Package strigo provides a comprehensive and flexible rate limiter for Go applications,
inspired by the popular Node.js package rate-limiter-flexible.

# Overview

StriGO implements various rate limiting algorithms with support for multiple storage
backends. It follows the design principles of rate-limiter-flexible with a simple,
intuitive API that provides detailed information about rate limit status.

# Key Features

   - Simple, intuitive API similar to rate-limiter-flexible
   - Multiple rate limiting strategies (Token Bucket, Leaky Bucket, Fixed Window, Sliding Window)
   - Flexible storage backends (Memory, Redis, Memcached)
   - Point-based system for variable cost operations
   - Detailed result information with standard HTTP headers
   - Framework-agnostic design
   - High performance with atomic operations
   - Modular project structure for maintainable code

# Basic Usage

Create a simple in-memory rate limiter:

	opts := &strigo.Options{
		Points:   5,  // 5 requests
		Duration: 10, // per 10 seconds
	}

	limiter, err := strigo.New(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer limiter.Close()

	// Consume 1 point
	result, err := limiter.Consume("user:123", 1)
	if err != nil {
		log.Fatal(err)
	}

	if result.Allowed {
		fmt.Printf("✅ Request allowed! Remaining: %d\n", result.RemainingPoints)
	} else {
		fmt.Printf("❌ Rate limited! Try again in %dms\n", result.MsBeforeNext)
	}

# Redis Storage

Use Redis for distributed rate limiting:

	import "github.com/redis/go-redis/v9"

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	opts := &strigo.Options{
		Points:      100, // 100 requests
		Duration:    60,  // per minute
		StoreClient: redisClient,
		KeyPrefix:   "myapp",
	}

	limiter, err := strigo.New(opts)

# Variable Point Consumption

Different operations can consume different amounts of points:

	limiter, _ := strigo.New(&strigo.Options{
		Points:   100, // 100 points total
		Duration: 60,  // per minute
	})

	// Light operation - 1 point
	result, _ := limiter.Consume("user:123", 1)

	// Heavy operation - 10 points
	result, _ := limiter.Consume("user:123", 10)

	// Very heavy operation - 25 points
	result, _ := limiter.Consume("user:123", 25)

# Recommended Project Structure

For maintainable applications, organize your rate limiters in separate files:

limiter.go - Define global rate limiter instances:

	var (
		ApiLimiter     *strigo.RateLimiter
		AuthLimiter    *strigo.RateLimiter
		UploadLimiter  *strigo.RateLimiter
		PremiumLimiter *strigo.RateLimiter
	)

	func InitializeLimiters() {
		redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

		var err error
		ApiLimiter, err = strigo.New(&strigo.Options{
			Points:      100,
			Duration:    60,
			KeyPrefix:   "api",
			StoreClient: redisClient,
		})
		if err != nil {
			log.Printf("⚠️  Redis not available, using memory: %v", err)
			ApiLimiter, _ = strigo.New(&strigo.Options{
				Points:    100,
				Duration:  60,
				KeyPrefix: "api",
			})
		}
		// ... initialize other limiters
	}

middleware.go - Create reusable middleware:

	// Generic HTTP middleware function
	func rateLimitMiddleware(limiter *strigo.RateLimiter, points ...int64) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			key := getUserKey(r)

			consumePoints := int64(1)
			if len(points) > 0 {
				consumePoints = points[0]
			}

			result, err := limiter.Consume(key, consumePoints)
			if err != nil {
				http.Error(w, "Rate limiter error", http.StatusInternalServerError)
				return
			}

			// Add standard rate limit headers
			headers := result.Headers()
			for name, value := range headers {
				w.Header().Set(name, value)
			}

			if !result.Allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, `{"error":"Rate limit exceeded","retryAfter":%d,"remaining":%d}`,
					result.MsBeforeNext/1000, result.RemainingPoints)
				return
			}

			// Continue to next handler
		}
	}

main.go - Use in your application:

	func main() {
		// Initialize rate limiters
		InitializeLimiters()

		// Create HTTP server with rate limiting
		http.HandleFunc("/api/users", rateLimitMiddleware(ApiLimiter)(getUsersHandler))
		http.HandleFunc("/api/premium", rateLimitMiddleware(PremiumLimiter, 1)(getPremiumHandler))
		http.HandleFunc("/api/analytics", rateLimitMiddleware(ApiLimiter, 5)(getAnalyticsHandler))
		http.HandleFunc("/auth/login", rateLimitMiddleware(AuthLimiter, 1)(loginHandler))
		http.HandleFunc("/upload", rateLimitMiddleware(UploadLimiter, 1)(uploadHandler))

		log.Fatal(http.ListenAndServe(":3000", nil))
	}

# Rate Limiting Strategies

The package supports multiple algorithms:

   - **TokenBucket** (default): Classic token bucket algorithm
   - **LeakyBucket**: Leaky bucket algorithm for smooth traffic
   - **FixedWindow**: Fixed time window counting
   - **SlidingWindow**: Sliding time window for more accurate limiting

# Storage Backends

   - **Memory**: Built-in in-memory storage (default)
   - **Redis**: Distributed rate limiting with Redis
   - **Memcached**: Distributed rate limiting with Memcached

# Result Object

The Consume method returns detailed information:

	type Result struct {
		MsBeforeNext      int64 // Milliseconds before next action can be done
		RemainingPoints   int64 // Number of remaining points in current duration
		ConsumedPoints    int64 // Number of consumed points in current duration
		IsFirstInDuration bool  // Whether the action is first in current duration
		TotalHits         int64 // Total points allowed in the duration
		Allowed           bool  // Whether the request was allowed
	}

The result also provides standard HTTP headers via the Headers() method:
   - X-RateLimit-Limit
   - X-RateLimit-Remaining
   - X-RateLimit-Reset
   - Retry-After (when limited)

# Additional Operations

Check rate limit status without consuming points:

	result, err := limiter.Get("user:123")

Reset rate limit for a key:

	err := limiter.Reset("user:123")

Block a key for specific duration:

	err := limiter.Block("user:123", 300) // 300 seconds

For more examples and detailed documentation, visit:
https://github.com/veyselaksin/strigo
*/
package strigo

## Installation

	go get github.com/veyselaksin/strigo/v2@v2.0.0

## Import

	import "github.com/veyselaksin/strigo/v2"
