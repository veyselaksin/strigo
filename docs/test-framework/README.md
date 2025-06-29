# Test Framework Overview

This test framework provides comprehensive testing capabilities for Redis and Memcached cache implementations with Go.

## Core Components

### 1. Test Helpers

Essential utilities and helper functions:

```go
func NewRedisClient() *redis.Client
func NewMemcachedClient() *memcache.Client
func CleanupRedis(t *testing.T, rdb *redis.Client)
func CleanupMemcached(t *testing.T, mc *memcache.Client)
```

### 2. Test Categories

#### Redis Tests (`tests/redis/`)

- **Basic Operations**: Set, get, delete, expiration
- **Performance Tests**: Sequential and concurrent benchmarks
- **Edge Cases**: Large keys, special characters, extreme values

#### Memcached Tests (`tests/memcached/`)

- **Basic Operations**: Set, get, delete, expiration
- **Performance Tests**: Sequential and concurrent benchmarks
- **Edge Cases**: Key limitations, connection scenarios

### 3. Test Structure

```
tests/
├── redis/
│   ├── basic_test.go       # Basic operations
│   ├── performance_test.go # Performance benchmarks
│   └── edge_cases_test.go  # Edge cases and limits
├── memcached/
│   ├── basic_test.go       # Basic operations
│   ├── performance_test.go # Performance benchmarks
│   └── edge_cases_test.go  # Edge cases and limits
└── helpers/
    └── test_helpers.go     # Test utilities
```

### 4. Environment Management

- Docker-based test environment (`docker/Dockerfile.test`)
- Configurable test parameters via environment variables
- Isolated test execution with container networking
- Performance benchmarking with detailed metrics

### 5. Running Tests

```bash
# Run all tests
go test ./tests/... -v

# Run specific backend tests
go test ./tests/redis/... -v
go test ./tests/memcached/... -v

# Run with Docker
docker build -t strigo-tests -f docker/Dockerfile.test .
docker run --rm --network host strigo-tests go test ./tests/... -v
```

### 6. Performance Benchmarks

Expected performance metrics:

- **Redis Concurrent**: 100K+ req/s
- **Redis Sequential**: 11K+ req/s
- **Memcached Concurrent**: 89K+ req/s
- **Memcached Sequential**: 11K+ req/s
