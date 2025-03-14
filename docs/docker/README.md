# Docker Test Environment

## Overview
Our Docker-based test environment provides isolated and reproducible test execution.

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

# Run specific test suite
docker-compose run --rm tests go test ./tests/redis/...
``` 