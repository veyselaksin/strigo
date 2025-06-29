package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/veyselaksin/strigo"
)

func main() {
	// Memory Example (no external storage required)
	memoryExample()

	fmt.Println("\n" + strings.Repeat("-", 50))

	// Redis Example
	redisExample()

	fmt.Println("\n" + strings.Repeat("-", 50))

	// Custom Points Example
	customPointsExample()
}

func memoryExample() {
	fmt.Println("📚 Memory Storage Example:")
	fmt.Println("Using in-memory storage (no Redis/Memcached required)")

	// Create options - similar to rate-limiter-flexible
	opts := &strigo.Options{
		Points:   5,  // 5 requests
		Duration: 10, // per 10 seconds
	}

	// Create rate limiter - similar to new RateLimiterMemory(opts)
	limiter, err := strigo.New(opts)
	if err != nil {
		log.Fatal("Failed to create limiter:", err)
	}
	defer limiter.Close()

	// Simulate requests
	userKey := "user:123"
	fmt.Printf("Rate limit: %d requests per %d seconds\n", opts.Points, opts.Duration)
	fmt.Printf("Testing key: %s\n\n", userKey)

	for i := 1; i <= 7; i++ {
		// Consume 1 point (default) - similar to rateLimiter.consume(key)
		result, err := limiter.Consume(userKey, 1)
		if err != nil {
			fmt.Printf("❌ Request %d: Error - %v\n", i, err)
			continue
		}

		status := "✅ ALLOWED"
		if !result.Allowed {
			status = "❌ BLOCKED"
		}

		fmt.Printf("Request %d: %s\n", i, status)
		fmt.Printf("  Remaining: %d, Consumed: %d, Reset in: %dms\n", 
			result.RemainingPoints, result.ConsumedPoints, result.MsBeforeNext)
		
		time.Sleep(500 * time.Millisecond)
	}
}

func redisExample() {
	fmt.Println("🗄️  Redis Storage Example:")
	fmt.Println("Using Redis for distributed rate limiting")

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Create options with Redis storage
	opts := &strigo.Options{
		Points:      10, // 10 requests  
		Duration:    60, // per minute
		StoreClient: redisClient,
		KeyPrefix:   "myapp",
	}

	limiter, err := strigo.New(opts)
	if err != nil {
		log.Printf("⚠️  Failed to create Redis limiter (Redis not available?): %v", err)
		return
	}
	defer limiter.Close()

	userKey := "api:user456"
	fmt.Printf("Rate limit: %d requests per %d seconds\n", opts.Points, opts.Duration)
	fmt.Printf("Testing key: %s\n\n", userKey)

	for i := 1; i <= 3; i++ {
		result, err := limiter.Consume(userKey, 1)
		if err != nil {
			fmt.Printf("❌ Request %d: Error - %v\n", i, err)
			continue
		}

		status := "✅ ALLOWED"
		if !result.Allowed {
			status = "❌ BLOCKED"
		}

		fmt.Printf("Request %d: %s (Remaining: %d)\n", 
			i, status, result.RemainingPoints)
	}
}

func customPointsExample() {
	fmt.Println("🎯 Custom Points Example:")
	fmt.Println("Different operations consume different amounts of points")

	opts := &strigo.Options{
		Points:   100, // 100 points
		Duration: 60,  // per minute
	}

	limiter, err := strigo.New(opts)
	if err != nil {
		log.Fatal("Failed to create limiter:", err)
	}
	defer limiter.Close()

	userKey := "api:premium-user"

	operations := []struct {
		name   string
		points int64
	}{
		{"👁️  View Profile", 1},
		{"✏️  Update Profile", 5},
		{"📤 Upload File", 10},
		{"🔍 Complex Search", 15},
		{"📊 Generate Report", 25},
	}

	fmt.Printf("Rate limit: %d points per %d seconds\n\n", opts.Points, opts.Duration)

	for _, op := range operations {
		result, err := limiter.Consume(userKey, op.points)
		if err != nil {
			fmt.Printf("❌ %s: Error - %v\n", op.name, err)
			continue
		}

		status := "✅ ALLOWED"
		if !result.Allowed {
			status = "❌ BLOCKED"
		}

		fmt.Printf("%s: %s (Cost: %d points, Remaining: %d)\n",
			op.name, status, op.points, result.RemainingPoints)
	}
}
