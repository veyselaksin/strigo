# Fiber Integration Test Documentation

## Redis Integration Tests

### Test Cases

#### 1. Cache Operations
```go
func TestFiberRedisIntegration(t *testing.T) {
    t.Run("Cache Operations", func(t *testing.T) {
        // Test POST
        req := httptest.NewRequest("POST", "/cache/test-key", 
            strings.NewReader("test-value"))
        resp, err := app.Test(req)
        assert.NoError(t, err)
        assert.Equal(t, 200, resp.StatusCode)
    })
}
```

## Memcached Integration Tests

### Test Cases

#### 1. Cache Operations
```go
func TestFiberMemcachedIntegration(t *testing.T) {
    t.Run("Cache Operations", func(t *testing.T) {
        // Test GET
        req := httptest.NewRequest("GET", "/cache/test-key", nil)
        resp, err = app.Test(req)
        assert.NoError(t, err)
        assert.Equal(t, 200, resp.StatusCode)
    })
}
```

## Running Tests
```bash
# Run all integration tests
go test ./tests/fiber/... -v

# Run specific integration test
go test ./tests/fiber/... -run TestFiberRedisIntegration -v
``` 