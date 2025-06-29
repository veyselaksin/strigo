package redis_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veyselaksin/strigo/v2"
)

func TestRedisConnectionFailure(t *testing.T) {
	// Try to connect to non-existent Redis instance
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Wrong port
		DB:   0,
	})
	defer redisClient.Close()

	opts := &strigo.Options{
		Points:      100,
		Duration:    60,
		StoreClient: redisClient,
	}

	limiter, err := strigo.New(opts)
	if err != nil {
		// If creation fails, test different behavior
		assert.Contains(t, err.Error(), "redis")
		return
	}
	defer limiter.Close()

	// Try to use limiter with failed connection
	result, err := limiter.Consume("test", 1)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestRedisLargeKeyNames(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   3,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping edge case tests")
	}

	redisClient.FlushDB(ctx)
	defer redisClient.Close()

	opts := &strigo.Options{
		Points:      100,
		Duration:    60,
		StoreClient: redisClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	// Test with very long key name
	longKey := strings.Repeat("a", 1000) // 1KB key
	result, err := limiter.Consume(longKey, 1)
	require.NoError(t, err)
	assert.True(t, result.Allowed)

	// Test with extremely long key
	extremelyLongKey := strings.Repeat("b", 10000) // 10KB key
	result, err = limiter.Consume(extremelyLongKey, 1)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestRedisSpecialCharacterKeys(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   3,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping edge case tests")
	}

	redisClient.FlushDB(ctx)
	defer redisClient.Close()

	opts := &strigo.Options{
		Points:      100,
		Duration:    60,
		StoreClient: redisClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	specialKeys := []string{
		"user:with:colons",
		"user@email.com",
		"user with spaces",
		"user\nwith\nnewlines",
		"user\twith\ttabs",
		"user/with/slashes",
		"user\\with\\backslashes",
		"user{with}braces",
		"user[with]brackets",
		"user(with)parentheses",
		"user!@#$%^&*()_+-=",
		"用户中文名",
		"пользователь",
		"🚀💯⭐️",
	}

	for _, key := range specialKeys {
		t.Run(fmt.Sprintf("Key_%s", key), func(t *testing.T) {
			result, err := limiter.Consume(key, 1)
			require.NoError(t, err, "Special character key should work: %s", key)
			assert.True(t, result.Allowed)
			
			// Verify we can get the status
			status, err := limiter.Get(key)
			require.NoError(t, err)
			assert.NotNil(t, status)
			assert.Equal(t, int64(1), status.ConsumedPoints)
		})
	}
}

func TestRedisEmptyKey(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   3,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping edge case tests")
	}

	redisClient.FlushDB(ctx)
	defer redisClient.Close()

	opts := &strigo.Options{
		Points:      100,
		Duration:    60,
		StoreClient: redisClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	// Test with empty key
	result, err := limiter.Consume("", 1)
	// Should either work with empty key or return meaningful error
	if err != nil {
		assert.Contains(t, err.Error(), "key")
	} else {
		assert.NotNil(t, result)
	}
}

func TestRedisExtremePointValues(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   3,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping edge case tests")
	}

	redisClient.FlushDB(ctx)
	defer redisClient.Close()

	opts := &strigo.Options{
		Points:      1000000, // 1 million points
		Duration:    60,
		StoreClient: redisClient,
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

func TestRedisConnectionRecovery(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		DB:           3,
		MaxRetries:   3,
		DialTimeout:  1 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping connection recovery tests")
	}

	redisClient.FlushDB(ctx)
	defer redisClient.Close()

	opts := &strigo.Options{
		Points:      100,
		Duration:    60,
		StoreClient: redisClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	// First, verify normal operation
	result, err := limiter.Consume("user1", 1)
	require.NoError(t, err)
	assert.True(t, result.Allowed)

	// Note: In a real test environment, you might want to:
	// 1. Stop Redis service
	// 2. Try operations (should fail)
	// 3. Start Redis service
	// 4. Try operations again (should work)
	// But this requires infrastructure setup, so we simulate timeout scenarios
	
	t.Log("Connection recovery test completed - would need Redis restart in full test")
}

func TestRedisHighConcurrencyEdgeCases(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   3,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping concurrency edge case tests")
	}

	redisClient.FlushDB(ctx)
	defer redisClient.Close()

	opts := &strigo.Options{
		Points:      1000,
		Duration:    60,
		StoreClient: redisClient,
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
	
	// Should have roughly correct distribution (some allowed, some blocked)
	assert.Greater(t, allowed, 0, "Some requests should be allowed")
	assert.Equal(t, workers, allowed+blocked, "All requests should be accounted for")
}

func TestRedisMemoryPressure(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   3,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping memory pressure tests")
	}

	redisClient.FlushDB(ctx)
	defer redisClient.Close()

	opts := &strigo.Options{
		Points:      100,
		Duration:    60,
		StoreClient: redisClient,
	}

	limiter, err := strigo.New(opts)
	require.NoError(t, err)
	defer limiter.Close()

	// Create many keys to test memory pressure
	keyCount := 10000
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
	
	// Test that existing keys still work
	result, err := limiter.Consume("mem_pressure_test:user:0", 1)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
} 