# Redis Test Documentation

## Basic Tests
These tests cover fundamental Redis operations and functionality.

### Test Cases

#### 1. Set and Get Operations
```go
func TestRedisBasicOperations(t *testing.T) {
    t.Run("Set and Get String", func(t *testing.T) {
        err := rdb.Set(ctx, "test-key", "test-value", 0).Err()
        assert.NoError(t, err)

        val, err := rdb.Get(ctx, "test-key").Result()
        assert.NoError(t, err)
        assert.Equal(t, "test-value", val)
    })
}
```

#### 2. Key Expiration
```go
t.Run("Key Expiration", func(t *testing.T) {
    err := rdb.Set(ctx, "expire-key", "value", 1*time.Second).Err()
    assert.NoError(t, err)

    time.Sleep(2 * time.Second)
    _, err = rdb.Get(ctx, "expire-key").Result()
    assert.Error(t, err)
})
```

## Advanced Tests

### Test Cases

#### 1. Hash Operations
```go
t.Run("Hash Operations", func(t *testing.T) {
    err := rdb.HSet(ctx, "user:1", map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
        "age":   "30",
    }).Err()
    assert.NoError(t, err)
})
```

#### 2. List Operations
```go
t.Run("List Operations", func(t *testing.T) {
    err := rdb.LPush(ctx, "mylist", "first", "second", "third").Err()
    assert.NoError(t, err)
})
```

## Running Tests
```bash
# Run all Redis tests
go test ./tests/redis/... -v

# Run specific test
go test ./tests/redis/... -run TestRedisBasicOperations -v
``` 