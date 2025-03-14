# Test Framework Overview

This test framework provides comprehensive testing capabilities for cache implementations using Redis and Memcached with Go and the Fiber framework.

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
- Basic Operations Tests
- Advanced Operations Tests
- Integration Tests
- Performance Tests

### 3. Environment Management
- Docker-based test environment
- Configurable test parameters
- Isolated test execution 