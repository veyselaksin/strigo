/*
Package strigo provides a comprehensive and flexible rate limiter for Go applications,
inspired by the popular Node.js package rate-limiter-flexible.

# Overview

StriGO implements four distinct rate limiting algorithms with proper academic implementations.
Each strategy has unique characteristics and use cases. The package supports multiple storage
backends and provides detailed information about rate limit status.

# Key Features

  - Simple, intuitive API similar to rate-limiter-flexible
  - Multiple rate limiting strategies with correct algorithmic implementations:
  - Token Bucket: Gradual refill with burst capability
  - Leaky Bucket: Constant drain rate with request queueing
  - Sliding Window: Precise timestamp tracking
  - Fixed Window: Counter reset at intervals
  - Flexible storage backends (Memory, Redis, Memcached)
  - Point-based system for variable cost operations
  - Detailed result information with standard HTTP headers
  - Framework-agnostic design
  - High performance with atomic operations
  - Modular project structure for maintainable code

# Rate Limiting Strategies

## Token Bucket (Default)

The token bucket algorithm maintains a bucket of tokens that refill at a constant rate.
Requests consume tokens, and the bucket allows burst capacity up to its limit.

	// Token bucket allows bursts followed by gradual recovery
	opts := &strigo.Options{
		Points:   10,           // 10 token capacity
		Duration: 60,           // Refill 10 tokens per minute
		Strategy: TokenBucket,  // Gradual refill algorithm
	}

Technical: Stores current token count, last refill timestamp, and refill rate.
Tokens are added continuously based on elapsed time since last refill.

## Leaky Bucket

The leaky bucket algorithm queues incoming requests and processes them at a constant
rate, providing smooth traffic flow regardless of arrival patterns.

	// Leaky bucket smooths traffic with constant processing rate
	opts := &strigo.Options{
		Points:   5,            // Queue capacity of 5 requests
		Duration: 30,           // Process 5 requests per 30 seconds
		Strategy: LeakyBucket,  // Constant drain algorithm
	}

Technical: Maintains a queue of pending requests that drain at a fixed rate.
Requests are processed at exactly 1 request per (Duration/Points) seconds.

## Sliding Window

The sliding window algorithm tracks individual request timestamps within a rolling
time window, providing precise rate limiting without boundary effects.

	// Sliding window provides precise limiting
	opts := &strigo.Options{
		Points:   100,            // 100 requests
		Duration: 3600,           // per hour (any 60-minute period)
		Strategy: SlidingWindow,  // Timestamp tracking algorithm
	}

Technical: Stores array of request timestamps, removes expired entries on each
request. Window slides continuously with current time.

## Fixed Window

The fixed window algorithm uses a simple counter that resets completely at regular
intervals aligned to calendar boundaries.

	// Fixed window with periodic resets
	opts := &strigo.Options{
		Points:   1000,          // 1000 requests
		Duration: 3600,          // per hour (top of each hour)
		Strategy: FixedWindow,   // Counter reset algorithm
	}

Technical: Single counter with TTL that resets to zero at fixed time boundaries.
Simplest implementation with lowest memory usage.

# Basic Usage

Create a simple in-memory rate limiter:

	opts := &strigo.Options{
		Points:   5,  // 5 requests
		Duration: 10, // per 10 seconds
		Strategy: TokenBucket, // Allow bursts with gradual refill
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
		Points:      100,              // 100 requests
		Duration:    60,               // per minute
		Strategy:    SlidingWindow,    // Precise limiting
		StoreClient: redisClient,
		KeyPrefix:   "myapp",
	}

	limiter, err := strigo.New(opts)

# Variable Point Consumption

Different operations can consume different amounts of points:

	limiter, _ := strigo.New(&strigo.Options{
		Points:   100,         // 100 points total
		Duration: 60,          // per minute
		Strategy: TokenBucket, // Allows burst consumption
	})

	// Light operation - 1 point
	result, _ := limiter.Consume("user:123", 1)

	// Heavy operation - 10 points
	result, _ := limiter.Consume("user:123", 10)

	// Very heavy operation - 25 points
	result, _ := limiter.Consume("user:123", 25)

# Strategy Behavior Comparison

Each strategy exhibits distinct behavior patterns:

Token Bucket:
- Allows immediate bursts up to capacity
- Gradual refill enables sustained usage
- Best for APIs allowing occasional spikes

Leaky Bucket:
- Queues excess requests for later processing
- Smooths irregular traffic patterns
- Best for backend services requiring steady load

Sliding Window:
- Precise rate limiting without window edge effects
- Higher memory usage due to timestamp storage
- Best when exact rate compliance is critical

Fixed Window:
- Simple counter with periodic resets
- Potential for double-rate at window boundaries
- Best for simple use cases with clear reset times

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
			Strategy:    TokenBucket,  // Allow API bursts
			KeyPrefix:   "api",
			StoreClient: redisClient,
		})
		if err != nil {
			log.Printf("⚠️  Redis not available, using memory: %v", err)
			ApiLimiter, _ = strigo.New(&strigo.Options{
				Points:    100,
				Duration:  60,
				Strategy:  TokenBucket,
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

		// Create HTTP server with different strategies
		http.HandleFunc("/api/users", rateLimitMiddleware(ApiLimiter)(getUsersHandler))
		http.HandleFunc("/api/premium", rateLimitMiddleware(PremiumLimiter, 1)(getPremiumHandler))
		http.HandleFunc("/api/analytics", rateLimitMiddleware(ApiLimiter, 5)(getAnalyticsHandler))
		http.HandleFunc("/auth/login", rateLimitMiddleware(AuthLimiter, 1)(loginHandler))
		http.HandleFunc("/upload", rateLimitMiddleware(UploadLimiter, 1)(uploadHandler))

		log.Fatal(http.ListenAndServe(":3000", nil))
	}

# Storage Backends

  - **Memory**: Built-in in-memory storage (default)
  - **Redis**: Distributed rate limiting with Redis
  - **Memcached**: Distributed rate limiting with Memcached

All storage backends support both simple counters (for Fixed Window) and complex
JSON objects (for Token Bucket, Leaky Bucket, and Sliding Window strategies).

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

# Performance Considerations

- **Token Bucket**: Low memory usage, efficient for most use cases
- **Leaky Bucket**: Medium memory usage, constant processing overhead
- **Sliding Window**: High memory usage (stores all timestamps), precise but expensive
- **Fixed Window**: Lowest memory usage, fastest performance

Choose the strategy based on your precision requirements and performance constraints.

For more examples and detailed documentation, visit:
https://github.com/veyselaksin/strigo

Installation:

	go get github.com/veyselaksin/strigo/v2@v2.0.0

Import:

	import "github.com/veyselaksin/strigo/v2"
*/
package strigo
