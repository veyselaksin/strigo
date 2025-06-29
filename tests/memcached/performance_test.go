package memcached_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veyselaksin/strigo"
)

func setupMemcachedForPerformance(t *testing.T) (*strigo.RateLimiter, func()) {
	memcachedClient := memcache.New("localhost:11211")

	// Test Memcached connection
	err := memcachedClient.Ping()
	if err != nil {
		t.Skip("Memcached not available, skipping performance tests")
	}

	// Clean test data
	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      1000,
		Duration:    60,
		KeyPrefix:   "perf_test",
		StoreClient: memcachedClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)

	cleanup := func() {
		memcachedClient.FlushAll()
		limiter.Close()
	}

	return limiter, cleanup
}

func TestMemcachedPerformanceSequential(t *testing.T) {
	limiter, cleanup := setupMemcachedForPerformance(t)
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
	
	// Memcached should be fast for sequential requests
	assert.Greater(t, rps, 100.0, "Sequential Memcached performance should be > 100 req/s")
}

func TestMemcachedPerformanceConcurrent(t *testing.T) {
	limiter, cleanup := setupMemcachedForPerformance(t)
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
	
	// Memcached should handle concurrent requests well
	assert.Greater(t, rps, 150.0, "Concurrent Memcached performance should be > 150 req/s")
	assert.Equal(t, requests, allowedCount, "All requests should be allowed (within limits)")
}

func TestMemcachedPerformanceVariablePoints(t *testing.T) {
	limiter, cleanup := setupMemcachedForPerformance(t)
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

func TestMemcachedPerformanceBurst(t *testing.T) {
	limiter, cleanup := setupMemcachedForPerformance(t)
	defer cleanup()

	// Test burst performance - many requests in short time
	burstSize := 500
	start := time.Now()

	for i := 0; i < burstSize; i++ {
		key := fmt.Sprintf("burst_user:%d", i%20) // 20 different users
		result, err := limiter.Consume(key, 1)
		require.NoError(t, err)
		
		// First requests should be allowed for each user
		if result.ConsumedPoints <= 50 { // Within burst allowance
			assert.True(t, result.Allowed, "Burst requests should be allowed initially")
		}
	}

	duration := time.Since(start)
	rps := float64(burstSize) / duration.Seconds()

	t.Logf("Burst Performance: %d requests in %v (%.2f req/s)", burstSize, duration, rps)
	assert.Greater(t, rps, 200.0, "Burst performance should be > 200 req/s")
}

func TestMemcachedPerformanceGetStatus(t *testing.T) {
	limiter, cleanup := setupMemcachedForPerformance(t)
	defer cleanup()

	// First, create some data
	keys := make([]string, 100)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("status_user:%d", i)
		keys[i] = key
		_, err := limiter.Consume(key, int64(i%10+1)) // Different point consumptions
		require.NoError(t, err)
	}

	// Test Get performance
	start := time.Now()
	for i := 0; i < 1000; i++ {
		key := keys[i%len(keys)]
		status, err := limiter.Get(key)
		require.NoError(t, err)
		assert.NotNil(t, status)
		assert.Greater(t, status.ConsumedPoints, int64(0))
	}
	duration := time.Since(start)

	rps := float64(1000) / duration.Seconds()
	t.Logf("Get Status Performance: 1000 gets in %v (%.2f gets/s)", duration, rps)
	assert.Greater(t, rps, 200.0, "Get status operations should be > 200 gets/s")
}

// Benchmark tests for Memcached
func BenchmarkMemcachedConsume(b *testing.B) {
	memcachedClient := memcache.New("localhost:11211")

	err := memcachedClient.Ping()
	if err != nil {
		b.Skip("Memcached not available, skipping benchmark")
	}

	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      10000,
		Duration:    60,
		KeyPrefix:   "bench",
		StoreClient: memcachedClient,
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

func BenchmarkMemcachedGet(b *testing.B) {
	memcachedClient := memcache.New("localhost:11211")

	err := memcachedClient.Ping()
	if err != nil {
		b.Skip("Memcached not available, skipping benchmark")
	}

	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      10000,
		Duration:    60,
		KeyPrefix:   "bench_get",
		StoreClient: memcachedClient,
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

func BenchmarkMemcachedReset(b *testing.B) {
	memcachedClient := memcache.New("localhost:11211")

	err := memcachedClient.Ping()
	if err != nil {
		b.Skip("Memcached not available, skipping benchmark")
	}

	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      1000,
		Duration:    60,
		KeyPrefix:   "bench_reset",
		StoreClient: memcachedClient,
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

func BenchmarkMemcachedMixedOperations(b *testing.B) {
	memcachedClient := memcache.New("localhost:11211")

	err := memcachedClient.Ping()
	if err != nil {
		b.Skip("Memcached not available, skipping benchmark")
	}

	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      5000,
		Duration:    60,
		KeyPrefix:   "bench_mixed",
		StoreClient: memcachedClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(b, err)
	defer limiter.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("user:%d", i%100)
			
			switch i % 4 {
			case 0, 1: // 50% consume operations
				limiter.Consume(key, 1)
			case 2: // 25% get operations
				limiter.Get(key)
			case 3: // 25% reset operations
				limiter.Reset(key)
			}
			i++
		}
	})
} 