# StriGO 🚀

[![Go Version](https://img.shields.io/github/go-mod/go-version/veyselaksin/StriGO)](https://go.dev/)
[![Version](https://img.shields.io/github/v/release/veyselaksin/StriGO?include_prereleases)](https://github.com/veyselaksin/StriGO/releases)
[![License](https://img.shields.io/github/license/veyselaksin/StriGO)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/coverage-87%25-green)](https://github.com/veyselaksin/StriGO/actions)

StriGO is a comprehensive testing framework for Redis and Memcached implementations using Go and the Fiber framework. ⚡️

## ✨ Features
- 🔄 Redis and Memcached testing utilities
- 🌐 Integration with Fiber framework
- 🐳 Docker-based test environment
- 🛠️ Easy-to-use test helpers

## 🚀 Quick Start

### 📋 Prerequisites
- 🔧 Go 1.22.3 or later
- 🐳 Docker and Docker Compose
- 📦 Redis
- 💾 Memcached

### 📥 Installation
```bash
go get github.com/veyselaksin/strigo
```

## 💡 Basic Usage

### Redis Tests
```go
func TestRedisBasic(t *testing.T) {
    // Initialize Redis client
    rdb := helpers.NewRedisClient()
    defer helpers.CleanupRedis(t, rdb)

    // Set a value
    err := rdb.Set(ctx, "test-key", "test-value", 0).Err()
    assert.NoError(t, err)

    // Get the value
    val, err := rdb.Get(ctx, "test-key").Result()
    assert.NoError(t, err)
    assert.Equal(t, "test-value", val)
}
```

### Memcached Tests
```go
func TestMemcachedBasic(t *testing.T) {
    // Initialize Memcached client
    mc := helpers.NewMemcachedClient()
    defer helpers.CleanupMemcached(t, mc)

    // Set a value
    err := mc.Set(&memcache.Item{
        Key:   "test-key",
        Value: []byte("test-value"),
    })
    assert.NoError(t, err)

    // Get the value
    item, err := mc.Get("test-key")
    assert.NoError(t, err)
    assert.Equal(t, []byte("test-value"), item.Value)
}
```

### Fiber Integration Tests
```go
func TestFiberIntegration(t *testing.T) {
    app := fiber.New()
    rdb := helpers.NewRedisClient()
    defer helpers.CleanupRedis(t, rdb)

    // Setup route with Redis
    app.Get("/cache/:key", func(c *fiber.Ctx) error {
        key := c.Params("key")
        val, err := rdb.Get(c.Context(), key).Result()
        if err == redis.Nil {
            return c.Status(404).SendString("Not found")
        }
        return c.SendString(val)
    })

    // Test the endpoint
    req := httptest.NewRequest("GET", "/cache/test-key", nil)
    resp, err := app.Test(req)
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

### Docker Tests
```bash
# Run all tests
docker compose -f docker/docker-compose.yml run --rm tests

# Run specific test suite
docker compose -f docker/docker-compose.yml run --rm tests go test ./tests/redis/... -v

# Run with coverage
docker compose -f docker/docker-compose.yml run --rm tests go test ./tests/... -coverprofile=coverage.out
```

## 📚 Documentation
For detailed documentation, please visit our [Documentation](https://veyselaksin.github.io/strigo/).

## 🤝 Contributing
Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md).

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
