# Memcached Test Documentation

## Basic Tests
These tests cover fundamental Memcached operations.

### Test Cases

#### 1. Set and Get Operations
```go
func TestMemcachedBasicOperations(t *testing.T) {
    t.Run("Set and Get", func(t *testing.T) {
        err := mc.Set(&memcache.Item{
            Key:   "test-key",
            Value: []byte("test-value"),
        })
        assert.NoError(t, err)
    })
}
```

#### 2. Delete Operations
```go
t.Run("Delete", func(t *testing.T) {
    err := mc.Delete("delete-key")
    assert.NoError(t, err)
})
```

## Advanced Tests

### Test Cases

#### 1. Multiple Set and Get
```go
t.Run("Multiple Set and Get", func(t *testing.T) {
    items := []*memcache.Item{
        {Key: "key1", Value: []byte("value1")},
        {Key: "key2", Value: []byte("value2")},
    }
})
```

#### 2. Compare And Swap
```go
t.Run("Compare And Swap", func(t *testing.T) {
    err := mc.CompareAndSwap(item)
    assert.NoError(t, err)
})
```

## Running Tests
```bash
# Run all Memcached tests
go test ./tests/memcached/... -v

# Run specific test
go test ./tests/memcached/... -run TestMemcachedBasicOperations -v
``` 