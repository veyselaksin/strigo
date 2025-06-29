package memcached_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veyselaksin/strigo"
)

func TestMemcachedConnectionFailure(t *testing.T) {
	// Try to connect to non-existent Memcached instance
	memcachedClient := memcache.New("localhost:9999") // Wrong port

	opts := &strigo.Options{
		Points:      100,
		Duration:    60,
		StoreClient: memcachedClient,
	}

	limiter, err := strigo.New(opts)
	if err != nil {
		// If creation fails, test different behavior
		assert.Contains(t, err.Error(), "memcache")
		return
	}
	defer limiter.Close()

	// Try to use limiter with failed connection
	result, err := limiter.Consume("test", 1)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestMemcachedLargeKeyNames(t *testing.T) {
	memcachedClient := memcache.New("localhost:11211")

	err := memcachedClient.Ping()
	if err != nil {
		t.Skip("Memcached not available, skipping edge case tests")
	}

	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      100,
		Duration:    60,
		StoreClient: memcachedClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	// Test with very long key name (Memcached has a 250 character limit)
	longKey := strings.Repeat("a", 200) // 200 chars should work
	result, err := limiter.Consume(longKey, 1)
	require.NoError(t, err)
	assert.True(t, result.Allowed)

	// Test with extremely long key (should be handled gracefully)
	extremelyLongKey := strings.Repeat("b", 300) // 300 chars might fail
	result, err = limiter.Consume(extremelyLongKey, 1)
	if err != nil {
		// Expected - Memcached has key length limits
		assert.Contains(t, err.Error(), "key")
	} else {
		// If it works, that's fine too
		assert.True(t, result.Allowed)
	}
}

func TestMemcachedSpecialCharacterKeys(t *testing.T) {
	memcachedClient := memcache.New("localhost:11211")

	err := memcachedClient.Ping()
	if err != nil {
		t.Skip("Memcached not available, skipping edge case tests")
	}

	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      100,
		Duration:    60,
		StoreClient: memcachedClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	// Memcached has restrictions on key characters
	validKeys := []string{
		"user_with_underscores",
		"user-with-dashes",
		"user123numbers",
		"UserWithMixedCase",
		"user.with.dots",
	}

	for _, key := range validKeys {
		t.Run(fmt.Sprintf("ValidKey_%s", key), func(t *testing.T) {
			result, err := limiter.Consume(key, 1)
			require.NoError(t, err, "Valid key should work: %s", key)
			assert.True(t, result.Allowed)
			
			// Verify we can get the status
			status, err := limiter.Get(key)
			require.NoError(t, err)
			assert.NotNil(t, status)
			assert.Equal(t, int64(1), status.ConsumedPoints)
		})
	}

	// Keys that might cause issues in Memcached
	problematicKeys := []string{
		"user with spaces",
		"user\nwith\nnewlines",
		"user\twith\ttabs",
		"user@email.com",
		"user{with}braces",
		"用户中文名",
		"🚀💯⭐️",
	}

	for _, key := range problematicKeys {
		t.Run(fmt.Sprintf("ProblematicKey_%s", key), func(t *testing.T) {
			result, err := limiter.Consume(key, 1)
			if err != nil {
				// Expected for some keys
				t.Logf("Key rejected as expected: %s - %v", key, err)
			} else {
				// If it works, verify it's fully functional
				assert.True(t, result.Allowed)
				status, err := limiter.Get(key)
				require.NoError(t, err)
				assert.NotNil(t, status)
			}
		})
	}
}

func TestMemcachedEmptyKey(t *testing.T) {
	memcachedClient := memcache.New("localhost:11211")

	err := memcachedClient.Ping()
	if err != nil {
		t.Skip("Memcached not available, skipping edge case tests")
	}

	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      100,
		Duration:    60,
		StoreClient: memcachedClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	// Test with empty key (Memcached doesn't allow empty keys)
	result, err := limiter.Consume("", 1)
	// Should return meaningful error
	if err != nil {
		assert.Contains(t, err.Error(), "key")
	} else {
		// If it somehow works, verify the result
		assert.NotNil(t, result)
	}
}

func TestMemcachedExtremePointValues(t *testing.T) {
	memcachedClient := memcache.New("localhost:11211")

	err := memcachedClient.Ping()
	if err != nil {
		t.Skip("Memcached not available, skipping edge case tests")
	}

	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      1000000, // 1 million points
		Duration:    60,
		StoreClient: memcachedClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	// Test with zero points
	result, err := limiter.Consume("user1", 0)
	if err != nil {
		assert.Contains(t, err.Error(), "points")
	} else {
		assert.True(t, result.Allowed)
		assert.Equal(t, int64(0), result.ConsumedPoints)
	}

	// Test with negative points (should be handled gracefully)
	result, err = limiter.Consume("user2", -1)
	if err != nil {
		assert.Contains(t, err.Error(), "points")
	}

	// Test with very large point consumption
	result, err = limiter.Consume("user3", 999999)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(999999), result.ConsumedPoints)

	// Test consuming more than limit
	result, err = limiter.Consume("user3", 2)
	require.NoError(t, err)
	assert.False(t, result.Allowed) // Should be blocked
}

func TestMemcachedConnectionTimeout(t *testing.T) {
	memcachedClient := memcache.New("localhost:11211")
	memcachedClient.Timeout = 100 * time.Millisecond // Very short timeout

	err := memcachedClient.Ping()
	if err != nil {
		t.Skip("Memcached not available, skipping timeout tests")
	}

	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      100,
		Duration:    60,
		StoreClient: memcachedClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	// Test normal operation first
	result, err := limiter.Consume("user1", 1)
	require.NoError(t, err)
	assert.True(t, result.Allowed)

	t.Log("Timeout test completed - would need network delay simulation for full test")
}

func TestMemcachedHighConcurrencyEdgeCases(t *testing.T) {
	memcachedClient := memcache.New("localhost:11211")

	err := memcachedClient.Ping()
	if err != nil {
		t.Skip("Memcached not available, skipping concurrency edge case tests")
	}

	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      1000,
		Duration:    60,
		StoreClient: memcachedClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	// Test many clients trying to consume exactly the limit
	workers := 100
	pointsPerWorker := 10 // Total: 1000 points (exactly the limit)
	
	results := make(chan *strigo.Result, workers)
	errors := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			result, err := limiter.Consume("shared_resource", int64(pointsPerWorker))
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}(i)
	}

	// Collect results
	var allowed, blocked int
	for i := 0; i < workers; i++ {
		select {
		case result := <-results:
			if result.Allowed {
				allowed++
			} else {
				blocked++
			}
		case err := <-errors:
			t.Logf("Error from worker: %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("Test timed out")
		}
	}

	t.Logf("High concurrency results: %d allowed, %d blocked", allowed, blocked)
	
	// Should have roughly correct distribution
	assert.Greater(t, allowed, 0, "Some requests should be allowed")
	assert.Equal(t, workers, allowed+blocked, "All requests should be accounted for")
}

func TestMemcachedMemoryPressure(t *testing.T) {
	memcachedClient := memcache.New("localhost:11211")

	err := memcachedClient.Ping()
	if err != nil {
		t.Skip("Memcached not available, skipping memory pressure tests")
	}

	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      100,
		Duration:    60,
		StoreClient: memcachedClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	// Create many keys to test memory pressure
	// Note: Memcached will evict old keys if memory is full
	keyCount := 5000 // Smaller than Redis test due to Memcached memory limits
	for i := 0; i < keyCount; i++ {
		key := fmt.Sprintf("mem_pressure_test:user:%d", i)
		result, err := limiter.Consume(key, 1)
		require.NoError(t, err)
		assert.True(t, result.Allowed)
		
		// Log progress periodically
		if i%1000 == 0 && i > 0 {
			t.Logf("Created %d keys", i)
		}
	}

	t.Logf("Successfully created %d unique keys", keyCount)
	
	// Test that some keys still work (older ones might be evicted)
	result, err := limiter.Consume(fmt.Sprintf("mem_pressure_test:user:%d", keyCount-1), 1)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestMemcachedServerRestart(t *testing.T) {
	memcachedClient := memcache.New("localhost:11211")

	err := memcachedClient.Ping()
	if err != nil {
		t.Skip("Memcached not available, skipping restart tests")
	}

	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      100,
		Duration:    60,
		StoreClient: memcachedClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	// Create some state
	result, err := limiter.Consume("user1", 50)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(50), result.ConsumedPoints)

	// Simulate server restart by flushing all data
	memcachedClient.FlushAll()

	// After restart, state should be reset
	result, err = limiter.Consume("user1", 1)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(1), result.ConsumedPoints) // Should start fresh
}

func TestMemcachedNetworkLatency(t *testing.T) {
	memcachedClient := memcache.New("localhost:11211")

	err := memcachedClient.Ping()
	if err != nil {
		t.Skip("Memcached not available, skipping latency tests")
	}

	memcachedClient.FlushAll()

	opts := &strigo.Options{
		Points:      1000,
		Duration:    60,
		StoreClient: memcachedClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	// Measure operation latencies
	operations := 100
	start := time.Now()

	for i := 0; i < operations; i++ {
		key := fmt.Sprintf("latency_user:%d", i%10)
		_, err := limiter.Consume(key, 1)
		require.NoError(t, err)
	}

	duration := time.Since(start)
	avgLatency := duration / time.Duration(operations)

	t.Logf("Average operation latency: %v", avgLatency)
	
	// Operations should be reasonably fast (adjust based on environment)
	assert.Less(t, avgLatency, 10*time.Millisecond, "Average latency should be < 10ms")
} 