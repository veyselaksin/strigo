package memcached

import (
	"testing"
	"time"

	"github.com/veyselaksin/strigo/v2/tests/helpers"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/stretchr/testify/assert"
)

func TestMemcachedBasicOperations(t *testing.T) {
	mc := helpers.NewMemcachedClient()
	defer helpers.CleanupMemcached(t, mc)

	t.Run("Set and Get", func(t *testing.T) {
		err := mc.Set(&memcache.Item{
			Key:   "test-key",
			Value: []byte("test-value"),
		})
		assert.NoError(t, err)

		item, err := mc.Get("test-key")
		assert.NoError(t, err)
		assert.Equal(t, []byte("test-value"), item.Value)
	})

	t.Run("Delete", func(t *testing.T) {
		// Set item
		err := mc.Set(&memcache.Item{
			Key:   "delete-key",
			Value: []byte("value"),
		})
		assert.NoError(t, err)

		// Delete item
		err = mc.Delete("delete-key")
		assert.NoError(t, err)

		// Verify deletion
		_, err = mc.Get("delete-key")
		assert.Equal(t, memcache.ErrCacheMiss, err)
	})

	t.Run("Expiration", func(t *testing.T) {
		err := mc.Set(&memcache.Item{
			Key:        "expire-key",
			Value:      []byte("value"),
			Expiration: 1, // 1 second
		})
		assert.NoError(t, err)

		time.Sleep(2 * time.Second)
		_, err = mc.Get("expire-key")
		assert.Equal(t, memcache.ErrCacheMiss, err)
	})
}
