package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/veyselaksin/strigo/v2"
)

func main() {
	fmt.Println("🚀 StriGO Strategy Comparison Demo")
	fmt.Println("Demonstrating different behaviors of each rate limiting algorithm")
	fmt.Println(strings.Repeat("=", 70))

	// Test each strategy with the same configuration
	testConfig := &strigo.Options{
		Points:   3, // Allow 3 requests
		Duration: 6, // per 6 seconds
	}

	strategies := []struct {
		name     string
		strategy strigo.Strategy
		desc     string
	}{
		{
			name:     "Token Bucket",
			strategy: strigo.TokenBucket,
			desc:     "Allows bursts, then gradual refill",
		},
		{
			name:     "Leaky Bucket", 
			strategy: strigo.LeakyBucket,
			desc:     "Queues requests, processes at constant rate",
		},
		{
			name:     "Sliding Window",
			strategy: strigo.SlidingWindow,
			desc:     "Tracks exact timestamps, precise limiting",
		},
		{
			name:     "Fixed Window",
			strategy: strigo.FixedWindow,
			desc:     "Simple counter, resets at intervals",
		},
	}

	for _, strategy := range strategies {
		fmt.Printf("\n📊 Testing %s Strategy\n", strategy.name)
		fmt.Printf("Description: %s\n", strategy.desc)
		fmt.Println(strings.Repeat("-", 50))
		
		testStrategy(strategy.name, strategy.strategy, testConfig)
		time.Sleep(1 * time.Second) // Brief pause between tests
	}

	fmt.Println("\n🎯 Advanced Behavior Demonstration")
	fmt.Println(strings.Repeat("=", 70))
	
	// Demonstrate burst behavior differences
	demonstrateBurstBehavior()
	
	// Demonstrate timing precision differences  
	demonstrateTimingPrecision()

	fmt.Println("\n✅ Strategy comparison complete!")
	fmt.Println("Each algorithm shows distinct behavior patterns.")
}

func testStrategy(name string, strategy strigo.Strategy, config *strigo.Options) {
	// Create limiter with specific strategy
	opts := *config // Copy config
	opts.Strategy = strategy
	
	limiter, err := strigo.New(&opts)
	if err != nil {
		log.Printf("Failed to create %s limiter: %v", name, err)
		return
	}
	defer limiter.Close()

	key := fmt.Sprintf("test:%s", strings.ToLower(strings.ReplaceAll(name, " ", "_")))
	
	// Test rapid requests (burst scenario)
	fmt.Printf("Rapid burst test (5 requests immediately):\n")
	for i := 1; i <= 5; i++ {
		result, err := limiter.Consume(key, 1)
		if err != nil {
			fmt.Printf("  Request %d: ERROR - %v\n", i, err)
			continue
		}
		
		status := "✅ ALLOWED"
		if !result.Allowed {
			status = fmt.Sprintf("❌ BLOCKED (retry in %dms)", result.MsBeforeNext)
		}
		
		fmt.Printf("  Request %d: %s (remaining: %d)\n", 
			i, status, result.RemainingPoints)
	}

	// Wait and test recovery
	fmt.Printf("\nWaiting 3 seconds for recovery...\n")
	time.Sleep(3 * time.Second)
	
	result, _ := limiter.Consume(key, 1)
	status := "✅ ALLOWED"
	if !result.Allowed {
		status = fmt.Sprintf("❌ BLOCKED (retry in %dms)", result.MsBeforeNext)
	}
	fmt.Printf("After 3s wait: %s (remaining: %d)\n", status, result.RemainingPoints)
}

func demonstrateBurstBehavior() {
	fmt.Println("\n🔥 Burst Behavior Comparison")
	fmt.Println("Testing how strategies handle sudden traffic spikes")
	fmt.Println(strings.Repeat("-", 50))

	config := &strigo.Options{
		Points:   5,  // 5 requests
		Duration: 10, // per 10 seconds
	}

	strategies := []strigo.Strategy{
		strigo.TokenBucket,
		strigo.FixedWindow,
	}

	for _, strategy := range strategies {
		opts := *config
		opts.Strategy = strategy
		
		limiter, _ := strigo.New(&opts)
		defer limiter.Close()

		strategyName := getStrategyName(strategy)
		fmt.Printf("\n%s - 10 rapid requests:\n", strategyName)
		
		allowed := 0
		blocked := 0
		
		for i := 1; i <= 10; i++ {
			result, _ := limiter.Consume(fmt.Sprintf("burst:%s", strategyName), 1)
			if result.Allowed {
				allowed++
				fmt.Printf("  ✅ Request %d allowed\n", i)
			} else {
				blocked++
				fmt.Printf("  ❌ Request %d blocked\n", i)
			}
		}
		
		fmt.Printf("Summary: %d allowed, %d blocked\n", allowed, blocked)
	}
}

func demonstrateTimingPrecision() {
	fmt.Println("\n⏰ Timing Precision Comparison")
	fmt.Println("Testing precision of window boundaries")
	fmt.Println(strings.Repeat("-", 50))

	// Test sliding vs fixed window behavior
	slidingOpts := &strigo.Options{
		Points:   2,
		Duration: 5, // 2 requests per 5 seconds
		Strategy: strigo.SlidingWindow,
	}
	
	fixedOpts := &strigo.Options{
		Points:   2,
		Duration: 5, // 2 requests per 5 seconds  
		Strategy: strigo.FixedWindow,
	}

	slidingLimiter, _ := strigo.New(slidingOpts)
	fixedLimiter, _ := strigo.New(fixedOpts)
	defer slidingLimiter.Close()
	defer fixedLimiter.Close()

	fmt.Println("Consuming 2 requests, waiting 3 seconds, then trying 2 more:")
	
	// Test sliding window
	fmt.Println("\nSliding Window:")
	slidingLimiter.Consume("timing:sliding", 2) // Consume 2 requests
	fmt.Println("  ✅ Consumed 2 requests")
	
	time.Sleep(3 * time.Second)
	fmt.Println("  ⏰ Waited 3 seconds")
	
	result, _ := slidingLimiter.Consume("timing:sliding", 2)
	if result.Allowed {
		fmt.Println("  ✅ 2 more requests allowed") 
	} else {
		fmt.Printf("  ❌ Blocked (window still has %d requests)\n", 
			slidingOpts.Points - result.RemainingPoints)
	}

	// Test fixed window  
	fmt.Println("\nFixed Window:")
	fixedLimiter.Consume("timing:fixed", 2) // Consume 2 requests
	fmt.Println("  ✅ Consumed 2 requests")
	
	time.Sleep(3 * time.Second) 
	fmt.Println("  ⏰ Waited 3 seconds")
	
	result, _ = fixedLimiter.Consume("timing:fixed", 2)
	if result.Allowed {
		fmt.Println("  ✅ 2 more requests allowed")
	} else {
		fmt.Printf("  ❌ Blocked (retry in %dms)\n", result.MsBeforeNext)
	}
}

func getStrategyName(strategy strigo.Strategy) string {
	switch strategy {
	case strigo.TokenBucket:
		return "Token Bucket"
	case strigo.LeakyBucket:
		return "Leaky Bucket" 
	case strigo.SlidingWindow:
		return "Sliding Window"
	case strigo.FixedWindow:
		return "Fixed Window"
	default:
		return "Unknown"
	}
} 