package fiber

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/veyselaksin/strigo/tests/helpers"
)

func TestFiberRedisIntegration(t *testing.T) {
	rdb := helpers.NewRedisClient()
	defer helpers.CleanupRedis(t, rdb)

	app := fiber.New()

	// Setup test routes
	app.Get("/cache/:key", func(c *fiber.Ctx) error {
		key := c.Params("key")
		val, err := rdb.Get(c.Context(), key).Result()
		if err != nil {
			return c.Status(404).SendString("Not found")
		}
		return c.SendString(val)
	})

	app.Post("/cache/:key", func(c *fiber.Ctx) error {
		key := c.Params("key")
		value := c.Body()
		err := rdb.Set(c.Context(), key, value, 0).Err()
		if err != nil {
			return c.Status(500).SendString("Error setting cache")
		}
		return c.SendString("OK")
	})

	t.Run("Cache Operations", func(t *testing.T) {
		// Test POST
		req := httptest.NewRequest("POST", "/cache/test-key", strings.NewReader("test-value"))
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		// Test GET
		req = httptest.NewRequest("GET", "/cache/test-key", nil)
		resp, err = app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "test-value", string(body))
	})
}
