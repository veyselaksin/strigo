package examples

import (
	"fmt"
	"log"
	"time"

	"github.com/veyselaksin/strigo/pkg/config"
	"github.com/veyselaksin/strigo/pkg/duration"
	"github.com/veyselaksin/strigo/pkg/limiter"
)

// RedisExample demonstrates basic Redis rate limiting
func RedisExample() {
	rateLimiter, err := limiter.NewLimiter(limiter.Config{
		Backend: limiter.Redis,
		Address: "localhost:6379",
		Rules: []limiter.RuleConfig{
			{
				Pattern:  "user-.*",
				Strategy: config.SlidingWindow,
				Period:   duration.MINUTELY,
				Limit:    1,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rateLimiter.Close()

	key := "<user_id or ip address>" // or whatever you want to use as a key
	if allowed := rateLimiter.Allow(key); allowed {
		fmt.Println("Request allowed for user:", key)
	} else {
		fmt.Println("Rate limit exceeded for user:", key)
	}
}

// CustomConfigExample demonstrates using custom configuration
func CustomConfigExample() {
	cfg := config.NewDefaultConfig()
	cfg.Strategy = config.SlidingWindow
	cfg.Limit = 1000
	cfg.Period = duration.MINUTELY
	cfg.Prefix = "myapp"
	cfg.BackendConfig.Password = "secret"
	cfg.BackendConfig.Database = 1

	// Use the config in your application...
}

// BatchRequestExample demonstrates handling multiple requests
func BatchRequestExample() {
	rateLimiter, err := limiter.NewLimiter(limiter.Config{
		Backend: limiter.Redis,
		Address: "localhost:6379",
		Rules: []limiter.RuleConfig{
			{
				Pattern:  "user-.*",
				Strategy: config.SlidingWindow,
				Period:   duration.SECONDLY,
				Limit:    5,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rateLimiter.Close()

	key := "<user_id or ip address>" // or whatever you want to use as a key
	for i := 0; i < 10; i++ {
		if allowed := rateLimiter.Allow(key); allowed {
			fmt.Printf("Request %d allowed for user: %s\n", i+1, key)
		} else {
			fmt.Printf("Request %d rate limit exceeded for user: %s\n", i+1, key)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// ResetExample demonstrates resetting rate limits
func ResetExample() {
	rateLimiter, err := limiter.NewLimiter(limiter.Config{
		Backend: limiter.Redis,
		Address: "localhost:6379",
		Rules: []limiter.RuleConfig{
			{
				Pattern:  "user-.*",
				Strategy: config.SlidingWindow,
				Period:   duration.MINUTELY,
				Limit:    2,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rateLimiter.Close()

	key := "<user_id or ip address>" // or whatever you want to use as a key

	// Use up the rate limit
	rateLimiter.Allow(key)
	rateLimiter.Allow(key)

	if !rateLimiter.Allow(key) {
		fmt.Println("Rate limit exceeded")
	}

	// Reset the rate limit
	if err := rateLimiter.Reset(key); err != nil {
		log.Fatal(err)
	}

	if rateLimiter.Allow(key) {
		fmt.Println("Request allowed after reset")
	}
}
