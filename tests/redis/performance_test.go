package redis_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veyselaksin/strigo/v2"
)

func setupRedisForPerformance(t *testing.T) (*strigo.RateLimiter, func()) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use DB 1 for tests
	})

	// Test Redis connection
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping performance tests")
	}

	// Clean test database
	redisClient.FlushDB(ctx)

	opts := &strigo.Options{
		Points:      1000,
		Duration:    60,
		KeyPrefix:   "perf_test",
		StoreClient: redisClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)

	cleanup := func() {
		redisClient.FlushDB(ctx)
		redisClient.Close()
		limiter.Close()
	}

	return limiter, cleanup
}

func TestRedisPerformanceSequential(t *testing.T) {
	limiter, cleanup := setupRedisForPerformance(t)
	defer cleanup()

	// Test sequential requests performance
	start := time.Now()
	requests := 1000

	for i := 0; i < requests; i++ {
		key := fmt.Sprintf("user:%d", i%10) // 10 different users
		result, err := limiter.Consume(key, 1)
		require.NoError(t, err)
		assert.True(t, result.Allowed)
	}

	duration := time.Since(start)
	rps := float64(requests) / duration.Seconds()

	t.Logf("Sequential Performance: %d requests in %v (%.2f req/s)", requests, duration, rps)
	
	// Should process at least 100 requests per second
	assert.Greater(t, rps, 100.0, "Sequential Redis performance should be > 100 req/s")
}

func TestRedisPerformanceConcurrent(t *testing.T) {
	limiter, cleanup := setupRedisForPerformance(t)
	defer cleanup()

	// Test concurrent requests performance
	requests := 1000
	concurrency := 50
	requestsPerWorker := requests / concurrency

	start := time.Now()
	var wg sync.WaitGroup
	errors := make(chan error, requests)
	results := make(chan *strigo.Result, requests)

	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < requestsPerWorker; i++ {
				key := fmt.Sprintf("worker:%d:user:%d", workerID, i%5)
				result, err := limiter.Consume(key, 1)
				if err != nil {
					errors <- err
					return
				}
				results <- result
			}
		}(worker)
	}

	wg.Wait()
	close(errors)
	close(results)

	duration := time.Since(start)
	rps := float64(requests) / duration.Seconds()

	// Check for errors
	var errorCount int
	for err := range errors {
		t.Logf("Error: %v", err)
		errorCount++
	}
	assert.Equal(t, 0, errorCount, "No errors should occur during concurrent access")

	// Check results
	var allowedCount int
	for result := range results {
		if result.Allowed {
			allowedCount++
		}
	}

	t.Logf("Concurrent Performance: %d requests in %v (%.2f req/s)", requests, duration, rps)
	t.Logf("Allowed requests: %d/%d", allowedCount, requests)
	
	// Should handle concurrent requests efficiently
	assert.Greater(t, rps, 200.0, "Concurrent Redis performance should be > 200 req/s")
	assert.Equal(t, requests, allowedCount, "All requests should be allowed (within limits)")
}

func TestRedisPerformanceVariablePoints(t *testing.T) {
	limiter, cleanup := setupRedisForPerformance(t)
	defer cleanup()

	// Test performance with variable point consumption
	operations := []struct {
		name   string
		points int64
		count  int
	}{
		{"light", 1, 300},
		{"medium", 5, 100},
		{"heavy", 10, 50},
		{"expensive", 25, 20},
	}

	start := time.Now()
	totalRequests := 0

	for _, op := range operations {
		for i := 0; i < op.count; i++ {
			key := fmt.Sprintf("var_test:%s:%d", op.name, i%10)
			result, err := limiter.Consume(key, op.points)
			require.NoError(t, err)
			totalRequests++
			
			// Most requests should be allowed (we have 1000 points per minute)
			if !result.Allowed {
				t.Logf("Request blocked: %s operation, remaining: %d", op.name, result.RemainingPoints)
			}
		}
	}

	duration := time.Since(start)
	rps := float64(totalRequests) / duration.Seconds()

	t.Logf("Variable Points Performance: %d operations in %v (%.2f op/s)", totalRequests, duration, rps)
	assert.Greater(t, rps, 50.0, "Variable points operations should be > 50 op/s")
}

func TestRedisPerformanceMemoryUsage(t *testing.T) {
	limiter, cleanup := setupRedisForPerformance(t)
	defer cleanup()

	// Create many unique keys to test memory usage
	uniqueKeys := 1000
	requestsPerKey := 10

	start := time.Now()
	for i := 0; i < uniqueKeys; i++ {
		key := fmt.Sprintf("memory_test:user:%d", i)
		for j := 0; j < requestsPerKey; j++ {
			result, err := limiter.Consume(key, 1)
			require.NoError(t, err)
			assert.True(t, result.Allowed) // Should be allowed (10 requests per key)
		}
	}
	duration := time.Since(start)

	totalRequests := uniqueKeys * requestsPerKey
	rps := float64(totalRequests) / duration.Seconds()

	t.Logf("Memory Usage Test: %d unique keys, %d total requests in %v (%.2f req/s)", 
		uniqueKeys, totalRequests, duration, rps)
	
	// Should handle many unique keys efficiently
	assert.Greater(t, rps, 100.0, "Should handle many unique keys efficiently")
}

// Benchmark tests
func BenchmarkRedisConsume(b *testing.B) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   2, // Use DB 2 for benchmarks
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		b.Skip("Redis not available, skipping benchmark")
	}

	redisClient.FlushDB(ctx)
	defer redisClient.Close()

	opts := &strigo.Options{
		Points:      10000,
		Duration:    60,
		KeyPrefix:   "bench",
		StoreClient: redisClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(b, err)
	defer limiter.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("user:%d", i%100)
			_, err := limiter.Consume(key, 1)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

func BenchmarkRedisGet(b *testing.B) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   2,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		b.Skip("Redis not available, skipping benchmark")
	}

	redisClient.FlushDB(ctx)
	defer redisClient.Close()

	opts := &strigo.Options{
		Points:      10000,
		Duration:    60,
		KeyPrefix:   "bench_get",
		StoreClient: redisClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(b, err)
	defer limiter.Close()

	// Pre-populate some data
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("user:%d", i)
		limiter.Consume(key, 1)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("user:%d", i%100)
			_, err := limiter.Get(key)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

func BenchmarkRedisReset(b *testing.B) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   2,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		b.Skip("Redis not available, skipping benchmark")
	}

	redisClient.FlushDB(ctx)
	defer redisClient.Close()

	opts := &strigo.Options{
		Points:      1000,
		Duration:    60,
		KeyPrefix:   "bench_reset",
		StoreClient: redisClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(b, err)
	defer limiter.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("user:%d", i%50)
			// First consume to create the key
			limiter.Consume(key, 1)
			// Then reset
			err := limiter.Reset(key)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
} 