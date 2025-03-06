package examples

import (
	"fmt"
	"log"
	"time"

	"github.com/veyselaksin/strigo"
)

// MemcachedExample demonstrates basic Memcached rate limiting
func MemcachedExample() {
	rateLimiter, err := strigo.NewLimiter(strigo.LimiterConfig{
		Backend: strigo.Memcached,
		Address: "localhost:11211",
		Rules: []strigo.RuleConfig{
			{
				Pattern:  "user-.*",
				Strategy: strigo.TokenBucket,
				Period:   strigo.SECONDLY,
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

// MemcachedBatchExample demonstrates handling multiple requests with Memcached
func MemcachedBatchExample() {
	rateLimiter, err := strigo.NewLimiter(strigo.LimiterConfig{
		Backend: strigo.Memcached,
		Address: "localhost:11211",
		Rules: []strigo.RuleConfig{
			{
				Pattern:  "user-.*",
				Strategy: strigo.TokenBucket,
				Period:   strigo.SECONDLY,
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
