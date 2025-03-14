# Fiber Framework Integration Tests

## Overview
Integration tests ensure proper interaction between the Fiber framework and cache implementations.

## Test Structure

### 1. Redis Integration
```go
func TestFiberRedisIntegration(t *testing.T) {
    app := fiber.New()
    // Test setup and implementation
}
```

### 2. Memcached Integration
```go
func TestFiberMemcachedIntegration(t *testing.T) {
    app := fiber.New()
    // Test setup and implementation
}
```

## Common Patterns
- Request/Response handling
- Cache middleware testing
- Error handling scenarios

## Test Environment Setup

### Prerequisites
- Fiber framework
- Redis/Memcached instances
- Testing environment configured

### Helper Functions
Common helper functions used in integration tests:
- `setupTestApp()`: Creates and configures test Fiber application
- `createTestClient()`: Sets up test HTTP client
- `cleanupTestData()`: Removes test data after each test

## Running Integration Tests

### Run All Integration Tests
```bash
go test ./tests/fiber -v
```

### Run Specific Integration Test
```bash
go test ./tests/fiber -run TestFiberRedisIntegration -v
```

### Run with Coverage
```bash
go test ./tests/fiber -v -cover
```

## Best Practices

### Test Structure
Each integration test should:
1. Setup test environment
2. Configure middleware
3. Define test routes
4. Execute requests
5. Verify responses
6. Cleanup resources

### Error Handling
Tests should verify:
- Successful cache operations
- Cache miss scenarios
- Error responses
- Middleware behavior
- Edge cases 