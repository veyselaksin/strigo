# Docker Configuration

## Overview
The Docker environment provides isolated testing capabilities for Redis and Memcached implementations.

## Configuration Files

### 1. `Dockerfile.test`
```dockerfile
FROM golang:1.22.3-alpine

WORKDIR /app
COPY . .

RUN go mod download
CMD ["go", "test", "./tests/...", "-v"]
```

### 2. `docker-compose.yml`
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
    environment:
      - REDIS_HOST=redis
      - MEMCACHED_HOST=memcached

  redis:
    image: redis:alpine
    ports:
      - "6379:6379"

  memcached:
    image: memcached:alpine
    ports:
      - "11211:11211"
```

## Environment Variables

| Variable        | Description               | Default    |
|----------------|---------------------------|------------|
| `REDIS_HOST`   | Redis server hostname      | `redis`    |
| `REDIS_PORT`   | Redis server port          | `6379`     |
| `MEMCACHED_HOST` | Memcached server hostname | `memcached` |
| `MEMCACHED_PORT` | Memcached server port     | `11211`    |

## Usage

```bash
# Run all tests
docker-compose run --rm tests

# Run specific test suite
docker-compose run --rm tests go test ./tests/redis/...
```
