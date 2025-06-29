package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/veyselaksin/strigo/v2/tests/helpers"
)

func TestRedisBasicOperations(t *testing.T) {
	rdb := helpers.NewRedisClient()
	defer helpers.CleanupRedis(t, rdb)
	ctx := context.Background()

	t.Run("Set and Get String", func(t *testing.T) {
		err := rdb.Set(ctx, "test-key", "test-value", 0).Err()
		assert.NoError(t, err)

		val, err := rdb.Get(ctx, "test-key").Result()
		assert.NoError(t, err)
		assert.Equal(t, "test-value", val)
	})

	t.Run("Key Expiration", func(t *testing.T) {
		err := rdb.Set(ctx, "expire-key", "value", 1*time.Second).Err()
		assert.NoError(t, err)

		time.Sleep(2 * time.Second)
		_, err = rdb.Get(ctx, "expire-key").Result()
		assert.Error(t, err)
	})

	t.Run("Delete Key", func(t *testing.T) {
		// Set key
		err := rdb.Set(ctx, "delete-key", "value", 0).Err()
		assert.NoError(t, err)

		// Delete key
		err = rdb.Del(ctx, "delete-key").Err()
		assert.NoError(t, err)

		// Verify deletion
		_, err = rdb.Get(ctx, "delete-key").Result()
		assert.Error(t, err)
	})
}
