# Comprehensive Guide to Redis Testing in Go

## Overview
These tests cover basic Redis operations:
- Set/Get operations
- Key expiration (TTL) checks
- Deletion operations
- Basic error handling

## Test Cases

### 1. Set/Get Operations
```go
func TestRedisSetGet(t *testing.T) {
    t.Run("Basic Set/Get", func(t *testing.T) {
        rdb := helpers.NewRedisClient()
        defer helpers.CleanupRedis(t, rdb)

        err := rdb.Set(ctx, "test-key", "test-value", 0).Err()
        assert.NoError(t, err)

        val, err := rdb.Get(ctx, "test-key").Result()
        assert.NoError(t, err)
        assert.Equal(t, "test-value", val)
    })
}
```

### 2. Key Expiration Check
```go
func TestRedisExpiration(t *testing.T) {
    t.Run("TTL Check", func(t *testing.T) {
        rdb := helpers.NewRedisClient()
        defer helpers.CleanupRedis(t, rdb)

        err := rdb.Set(ctx, "expire-key", "value", time.Second).Err()
        assert.NoError(t, err)

        time.Sleep(2 * time.Second)
        _, err = rdb.Get(ctx, "expire-key").Result()
        assert.Equal(t, redis.Nil, err)
    })
}
```

### 3. Deletion Operations
```go
func TestRedisDelete(t *testing.T) {
    t.Run("Delete Key", func(t *testing.T) {
        rdb := helpers.NewRedisClient()
        defer helpers.CleanupRedis(t, rdb)

        // Create the key
        err := rdb.Set(ctx, "delete-key", "value", 0).Err()
        assert.NoError(t, err)

        // Delete the key
        err = rdb.Del(ctx, "delete-key").Err()
        assert.NoError(t, err)

        // Verify deletion
        _, err = rdb.Get(ctx, "delete-key").Result()
        assert.Equal(t, redis.Nil, err)
    })
}
```

## Running Tests

### Run All Basic Tests
```bash
go test ./tests/redis -run TestRedisBasic -v
```

### Run a Specific Test
```bash
go test ./tests/redis -run TestRedisSetGet -v
```

### Run with Coverage Report
```bash
go test ./tests/redis -run TestRedisBasic -v -cover
```

## Test Environment Setup

### Requirements
- A running Redis server (default: localhost:6379)
- Go test environment properly configured

### Helper Functions
The tests use common helper functions from the `helpers` package:
- `NewRedisClient()`: Creates a new Redis client
- `CleanupRedis()`: Cleans up Redis data after tests

---

# Fiber Framework Integration Tests

## Overview
These tests verify the correct operation of caching systems with the Fiber framework.

## Test Structure

### 1. Redis Integration
```go
func TestFiberRedisIntegration(t *testing.T) {
    app := fiber.New()
    // Test setup and application
}
```

### 2. Memcached Integration
```go
func TestFiberMemcachedIntegration(t *testing.T) {
    app := fiber.New()
    // Test setup and application
}
```

## Common Testing Patterns
- Request/Response handling
- Cache middleware testing
- Testing error scenarios

---

# Docker Test Environment

## Overview
Provides an isolated and reproducible test environment using Docker.

## Components

### 1. Services
- Redis container
- Memcached container
- Test runner container

### 2. Configuration
```yaml
version: '3.8'
services:
  tests:
    build:
      context: ..
      dockerfile: docker/Dockerfile.test
    depends_on:
      - redis
      - memcached
```

### 3. Usage
```bash
# Run all tests
./scripts/run-tests.sh

# Run a specific test package
docker-compose run --rm tests go test ./tests/redis/...
```

---

# Best Practices for Writing Tests

## General Guidelines

### 1. Test Organization
- Group related tests into sub-tests
- Use descriptive test names
- Follow the AAA (Arrange, Act, Assert) pattern

### 2. Independent Tests
```go
func TestIndependent(t *testing.T) {
    // Setup
    cleanup := setUp()
    defer cleanup()

    // Execute test
}
```

### 3. Error Handling
```go
func TestErrorCases(t *testing.T) {
    t.Run("Invalid Input", func(t *testing.T) {
        _, err := processInput("")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid input")
    })
}
```

## Common Testing Patterns

### 1. Table-Driven Tests
```go
func TestTableDriven(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "TEST",
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := processInput(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 2. Setup and Cleanup
```go
func TestWithSetup(t *testing.T) {
    // Setup
    rdb := helpers.NewRedisClient()
    defer helpers.CleanupRedis(t, rdb)

    // Test cases
    t.Run("first test", func(t *testing.T) {
        // Test implementation
    })

    t.Run("second test", func(t *testing.T) {
        // Test implementation
    })
}
```

