# Running Tests with Docker

## Common Commands

### Running All Tests
```bash
docker compose -f docker/docker-compose.yml run --rm tests
```

### Running Specific Test File
```bash
docker compose -f docker/docker-compose.yml run --rm tests go test ./tests/redis/basic_test.go -v
```

### Running Specific Test Function
```bash
docker compose -f docker/docker-compose.yml run --rm tests go test ./tests/... -run TestRedisBasicOperations -v
```

### Managing Services

#### Start Application
```bash
docker compose -f docker/docker-compose.yml up app
```

#### Start All Services in Background
```bash
docker compose -f docker/docker-compose.yml up -d
```

#### Stop All Services
```bash
docker compose -f docker/docker-compose.yml down
```

### Viewing Logs

#### View All Logs
```bash
docker compose -f docker/docker-compose.yml logs -f
```

#### View Test Service Logs
```bash
docker compose -f docker/docker-compose.yml logs -f tests
```

## Command Options Explained

| Option | Description |
|--------|-------------|
| `-f docker/docker-compose.yml` | Specifies the compose file location |
| `--rm` | Removes container after execution |
| `-v` | Enables verbose output for tests |
| `-d` | Runs services in background (detached mode) |
| `-f` (in logs) | Follows log output |

## Best Practices

1. Always use `--rm` when running tests to avoid leftover containers
2. Use `-d` when starting services that need to run in the background
3. Use verbose mode (`-v`) for detailed test output
4. Clean up resources using `down` command after testing
5. Use logs to debug test failures

## Examples

### Running a Test Suite with Custom Parameters
```bash
docker compose -f docker/docker-compose.yml run --rm tests go test ./tests/... -v -timeout 30s
```

### Running Tests with Race Detection
```bash
docker compose -f docker/docker-compose.yml run --rm tests go test -race ./tests/...
```

### Running Tests with Coverage
```bash
docker compose -f docker/docker-compose.yml run --rm tests go test ./tests/... -coverprofile=coverage.out
```

### Cleaning Up and Running Fresh Tests
```bash
docker compose -f docker/docker-compose.yml down --volumes
docker compose -f docker/docker-compose.yml run --rm tests
```

This documentation:
- Uses the new `docker compose` syntax instead of `docker-compose`
- Provides clear examples for each command
- Explains command options
- Includes best practices
- Shows advanced usage examples 