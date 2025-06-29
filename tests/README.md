# StriGO Test Suite

This directory contains comprehensive test suites for the StriGO rate limiter, covering Redis and Memcached backends with extensive performance and edge case testing.

## 📁 Test Structure

```
tests/
├── redis/                   # Redis backend tests
│   ├── basic_test.go       # Basic operations (set, get, delete, expiration)
│   ├── performance_test.go # Performance benchmarks and load testing
│   └── edge_cases_test.go  # Edge cases, limits, and special scenarios
├── memcached/              # Memcached backend tests
│   ├── basic_test.go       # Basic operations (set, get, delete, expiration)
│   ├── performance_test.go # Performance benchmarks and load testing
│   └── edge_cases_test.go  # Edge cases, limits, and special scenarios
└── helpers/                # Test utilities and helper functions
    └── test_helpers.go     # Common test functions and utilities
```

## 🚀 Running Tests

### Prerequisites

Ensure you have Redis and Memcached services running:

```bash
# Start Redis (default port 6379)
docker run -d -p 6379:6379 redis:alpine

# Start Memcached (default port 11211)
docker run -d -p 11211:11211 memcached:alpine
```

### Test Commands

```bash
# Run all tests
go test ./tests/... -v

# Run specific backend tests
go test ./tests/redis/... -v
go test ./tests/memcached/... -v

# Run specific test files
go test ./tests/redis/basic_test.go -v
go test ./tests/redis/performance_test.go -v

# Run performance benchmarks
go test ./tests/redis/performance_test.go -bench=. -v
go test ./tests/memcached/performance_test.go -bench=. -v

# Run tests with short flag (skips slow tests)
go test ./tests/... -short -v
```

## 🐳 Docker Testing

For isolated testing environments:

```bash
# Build test image
cd docker
docker build -t strigo-tests -f Dockerfile.test ..

# Run tests with Docker
docker run --rm --network host \
    -e REDIS_HOST=localhost \
    -e MEMCACHED_HOST=localhost \
    strigo-tests go test ./tests/... -v
```

## 📊 Test Categories

### Basic Operations Tests

- **Set and Get**: Basic storage and retrieval
- **Key Expiration**: Time-based expiration testing
- **Delete Operations**: Key deletion and cleanup
- **Connection Handling**: Connection failure scenarios

### Performance Tests

- **Sequential Performance**: Single-threaded operation speed
- **Concurrent Performance**: Multi-threaded operation speed
- **Variable Point Consumption**: Different point costs per operation
- **Memory Usage**: Large dataset handling (10K+ keys)
- **Burst Performance**: High-frequency request handling

### Edge Cases Tests

- **Large Key Names**: Keys approaching size limits
- **Special Characters**: Unicode, symbols, and special characters
- **Empty Keys**: Edge case handling for empty key names
- **Extreme Point Values**: Testing with very large point values
- **High Concurrency**: Stress testing with 100+ concurrent operations
- **Memory Pressure**: Testing with large numbers of keys

## 🎯 Performance Benchmarks

Expected performance metrics from recent Docker tests:

### Redis Performance

- **Concurrent**: 100,000+ req/s ⚡️
- **Sequential**: 11,000+ req/s
- **Variable Points**: 12,000+ op/s
- **Memory Usage**: 12,000+ req/s (10K keys)

### Memcached Performance

- **Concurrent**: 89,000+ req/s ⚡️
- **Sequential**: 11,000+ req/s
- **Variable Points**: 11,000+ op/s
- **Burst**: 11,500+ req/s
- **Get Status**: 22,000+ gets/s

## ✅ Test Coverage

The test suite covers:

- ✅ **Basic Operations**: All CRUD operations
- ✅ **Performance Metrics**: Speed and throughput validation
- ✅ **Edge Cases**: Boundary conditions and error scenarios
- ✅ **Connection Handling**: Network failure and recovery
- ✅ **Memory Management**: Large dataset operations
- ✅ **Concurrent Safety**: Thread-safe operation verification
- ✅ **Backend-Specific**: Redis vs Memcached differences

## 🔧 Test Configuration

Tests can be configured via environment variables:

```bash
# Redis configuration
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Memcached configuration
export MEMCACHED_HOST=localhost
export MEMCACHED_PORT=11211

# Test settings
export TEST_TIMEOUT=30s
export TEST_VERBOSE=true
```

## 📝 Writing New Tests

When adding new tests, follow these patterns:

```go
func TestNewFeature(t *testing.T) {
    // Setup
    client := helpers.NewRedisClient()
    defer helpers.CleanupRedis(t, client)

    limiter, err := strigo.New(&strigo.Options{
        Points:      10,
        Duration:    60,
        StoreClient: client,
    })
    assert.NoError(t, err)
    defer limiter.Close()

    // Test logic
    result, err := limiter.Consume("test-key", 1)
    assert.NoError(t, err)
    assert.True(t, result.Allowed)

    // Assertions
    assert.Equal(t, int64(9), result.RemainingPoints)
}
```

## 🎉 Test Results

All tests are designed to be:

- **Fast**: Most tests complete in milliseconds
- **Reliable**: Consistent results across environments
- **Isolated**: No dependencies between test cases
- **Comprehensive**: Cover normal and edge cases
- **Informative**: Clear error messages and debugging info

For detailed test documentation, see [Test Framework docs](../docs/test-framework/README.md).
