package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/veyselaksin/strigo/tests/helpers"
)

func TestRedisAdvancedOperations(t *testing.T) {
	rdb := helpers.NewRedisClient()
	defer helpers.CleanupRedis(t, rdb)
	ctx := context.Background()

	t.Run("Hash Operations", func(t *testing.T) {
		// Set hash fields
		err := rdb.HSet(ctx, "user:1", map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   "30",
		}).Err()
		assert.NoError(t, err)

		// Get specific field
		name, err := rdb.HGet(ctx, "user:1", "name").Result()
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", name)

		// Get all fields
		fields, err := rdb.HGetAll(ctx, "user:1").Result()
		assert.NoError(t, err)
		assert.Equal(t, 3, len(fields))
	})

	t.Run("List Operations", func(t *testing.T) {
		// Push elements
		err := rdb.LPush(ctx, "mylist", "first", "second", "third").Err()
		assert.NoError(t, err)

		// Get range
		elements, err := rdb.LRange(ctx, "mylist", 0, -1).Result()
		assert.NoError(t, err)
		assert.Equal(t, 3, len(elements))

		// Pop element
		val, err := rdb.LPop(ctx, "mylist").Result()
		assert.NoError(t, err)
		assert.Equal(t, "third", val)
	})

	t.Run("Set Operations", func(t *testing.T) {
		// Add members
		err := rdb.SAdd(ctx, "myset", "member1", "member2", "member3").Err()
		assert.NoError(t, err)

		// Check membership
		exists, err := rdb.SIsMember(ctx, "myset", "member1").Result()
		assert.NoError(t, err)
		assert.True(t, exists)

		// Get all members
		members, err := rdb.SMembers(ctx, "myset").Result()
		assert.NoError(t, err)
		assert.Equal(t, 3, len(members))
	})
}
