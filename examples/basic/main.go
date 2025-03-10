package main

import (
	"fmt"
	"log"
	"time"

	"github.com/veyselaksin/strigo"
)

func main() {
	// Redis Example
	redisExample()

	fmt.Println("\n--------------------------------")

	// Memcached Example
	memcachedExample()
}

func redisExample() {
	fmt.Println("Redis Example:")

	cfg := strigo.LimiterConfig{
		Backend: strigo.Redis,
		Address: "localhost:6379",
		Rules: []strigo.RuleConfig{
			{
				Pattern:  "api_requests",
				Period:   strigo.MINUTELY,
				Limit:    5,
				Strategy: strigo.TokenBucket,
			},
		},
		Prefix: "redis_example",
	}

	// Create Redis-based rate limiter
	redisLimiter, err := strigo.NewLimiter(cfg)
	if err != nil {
		log.Fatal("Failed to create Redis limiter:", err)
	}
	defer redisLimiter.Close()

	// Simulate requests
	key := buildKey(cfg.Prefix, cfg.Rules[0].Pattern, "user123")
	for i := 1; i <= 7; i++ {
		allowed := redisLimiter.Allow(key)
		fmt.Printf("Request %d: %v\n, key: %s\n", i, allowed, key)
		time.Sleep(time.Millisecond * 100) // Small delay between requests
	}
}

func memcachedExample() {
	fmt.Println("Memcached Example:")

	cfg := strigo.LimiterConfig{
		Backend: strigo.Memcached,
		Address: "localhost:11211",
		Rules: []strigo.RuleConfig{
			{
				Pattern:  "api_requests",
				Period:   strigo.MINUTELY,
				Limit:    5,
				Strategy: strigo.TokenBucket,
			},
		},
		Prefix: "memcached_example",
	}

	// Create Memcached-based rate limiter
	memcachedLimiter, err := strigo.NewLimiter(cfg)
	if err != nil {
		log.Fatal("Failed to create Memcached limiter:", err)
	}
	defer memcachedLimiter.Close()

	// Simulate requests
	key := buildKey(cfg.Prefix, cfg.Rules[0].Pattern, "user123")
	for i := 1; i <= 7; i++ {
		allowed := memcachedLimiter.Allow(key)
		fmt.Printf("Request %d: %v\n, key: %s\n", i, allowed, key)
		time.Sleep(time.Millisecond * 100) // Small delay between requests
	}
}

func buildKey(prefix string, pattern string, userID string) string {
	return prefix + ":" + pattern + ":" + userID
}
