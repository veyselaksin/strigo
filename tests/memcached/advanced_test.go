package memcached

import (
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/stretchr/testify/assert"
	"github.com/veyselaksin/strigo/tests/helpers"
)

func TestMemcachedAdvancedOperations(t *testing.T) {
	mc := helpers.NewMemcachedClient()
	defer helpers.CleanupMemcached(t, mc)

	t.Run("Multiple Set and Get", func(t *testing.T) {
		items := []*memcache.Item{
			{Key: "key1", Value: []byte("value1")},
			{Key: "key2", Value: []byte("value2")},
			{Key: "key3", Value: []byte("value3")},
		}

		// Set multiple items
		for _, item := range items {
			err := mc.Set(item)
			assert.NoError(t, err)
		}

		// Get multiple items
		keys := []string{"key1", "key2", "key3"}
		results, err := mc.GetMulti(keys)
		assert.NoError(t, err)
		assert.Equal(t, len(keys), len(results))
	})

	t.Run("Compare And Swap", func(t *testing.T) {
		// Initial set
		err := mc.Set(&memcache.Item{
			Key:   "cas-key",
			Value: []byte("initial-value"),
		})
		assert.NoError(t, err)

		// Get item with CAS
		item, err := mc.Get("cas-key")
		assert.NoError(t, err)

		// Perform CAS
		item.Value = []byte("new-value")
		err = mc.CompareAndSwap(item)
		assert.NoError(t, err)

		// Verify new value
		updated, err := mc.Get("cas-key")
		assert.NoError(t, err)
		assert.Equal(t, []byte("new-value"), updated.Value)
	})

	t.Run("Increment and Decrement", func(t *testing.T) {
		// Set initial counter
		err := mc.Set(&memcache.Item{
			Key:   "counter",
			Value: []byte("10"),
		})
		assert.NoError(t, err)

		// Increment
		newVal, err := mc.Increment("counter", 5)
		assert.NoError(t, err)
		assert.Equal(t, uint64(15), newVal)

		// Decrement
		newVal, err = mc.Decrement("counter", 3)
		assert.NoError(t, err)
		assert.Equal(t, uint64(12), newVal)
	})
}
