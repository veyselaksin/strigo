# StriGO

[![Tests](https://github.com/veyselaksin/StriGO/actions/workflows/tests.yml/badge.svg)](https://github.com/veyselaksin/StriGO/actions)
[![Go Coverage](https://github.com/veyselaksin/StriGO/wiki/coverage.svg)](https://raw.githubusercontent.com/veyselaksin/StriGO/main/coverage.out)

StriGO is a comprehensive testing framework for Redis and Memcached implementations using Go and the Fiber framework.

## Features
- Redis and Memcached testing utilities
- Integration with Fiber framework
- Docker-based test environment
- Comprehensive test coverage
- Easy-to-use test helpers

## Quick Start

### Prerequisites
- Go 1.22.3 or later
- Docker and Docker Compose
- Redis
- Memcached

### Installation
```bash
go get github.com/veyselaksin/StriGO
```

### Running Tests
```bash
# Run all tests
go test ./tests/... -v

# Run specific tests
go test ./tests/redis/... -v
go test ./tests/memcached/... -v
```

### Docker Support
```bash
# Run tests in Docker
docker compose -f docker/docker-compose.yml run --rm tests
```

## Documentation
For detailed documentation, please visit our [Documentation](docs/README.md).

## Contributing
Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md).

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

# StriGo 🦉

**StriGo** is a high-performance rate limiter for Go, designed to work seamlessly with Redis, Memcached, and Dragonfly. It provides efficient and scalable request limiting to protect your applications from abuse and ensure fair resource distribution.

[![Go Reference](https://pkg.go.dev/badge/github.com/veyselaksin/strigo.svg)](https://pkg.go.dev/github.com/veyselaksin/strigo)
[![Go Report Card](https://goreportcard.com/badge/github.com/veyselaksin/strigo)](https://goreportcard.com/report/github.com/veyselaksin/strigo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features 🚀

- 🔥 **High-performance**: Optimized for speed and efficiency
- 🔄 **Multiple Storage Backends**: 
  - Redis
  - Memcached
  - Dragonfly
- 🛡 **Advanced Rate Limiting Strategies**:
  - Token Bucket (Default)
  - Leaky Bucket
  - Fixed Window
  - Sliding Window
- ⚡ **Flexible Time Windows**:
  - Per Second
  - Per Minute
  - Per Hour
  - Per Day
  - Per Week
  - Per Month
  - Per Year
- 🌐 **Framework Integration**:
  - Fiber Framework Support
  - Easy to extend for other frameworks
- 📦 **Developer Friendly**:
  - Simple API
  - Comprehensive Documentation
  - Type-safe Configuration

## Documentation 📚

Visit our [Documentation Site](https://veyselaksin.github.io/StriGO) for:
- Getting Started Guide
- API Reference
- Advanced Usage Examples
- Best Practices

## Installation 📦

```sh
go get github.com/veyselaksin/strigo
```

## Quick Start ⚡

### Basic Usage with Redis

```go
package main

import (
    "log"
    "github.com/veyselaksin/strigo"
    "github.com/gofiber/fiber/v2"
)

func main() {
    // Create rate limiter
    limiter, err := strigo.NewLimiter(strigo.Config{
        Backend: strigo.Redis,
        Address: "localhost:6379",
        Strategy: strigo.TokenBucket,
        Period: strigo.MINUTELY,
        Limit: 100,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer limiter.Close()

    // Use in your application
    if limiter.Allow("user-123") {
        // Handle request
    } else {
        // Rate limit exceeded
    }
}
```

### Fiber Middleware Integration

```go
app := fiber.New()

// Create rate limiter manager
manager := strigo.NewManager(strigo.Redis, "localhost:6379")
defer manager.Close()

// Apply rate limiting middleware
app.Use(fiberMiddleware.RateLimitHandler(manager, func(c *fiber.Ctx) []strigo.RuleConfig {
    return []strigo.RuleConfig{
        {
            Pattern:  "api_limit",
            Strategy: strigo.TokenBucket,
            Period:   strigo.MINUTELY,
            Limit:    100,
        },
    }
}))
```

## Configuration Options 🔧

```go
type Config struct {
    Strategy      Strategy      // Rate limiting algorithm
    Period        Period        // Time window
    Limit         int64        // Maximum requests
    Backend       Backend      // Storage backend
    Address       string       // Backend address
    Prefix        string       // Key prefix
}
```

## Contributing 🤝

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License 📜

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support 💬

- Create an [Issue](https://github.com/veyselaksin/strigo/issues)
- Send an [Email](mailto:veyselaksn@gmail.com)

---

Made with ❤️ by [Veysel Aksin](https://github.com/veyselaksin)

# Cache Testing Project

This project demonstrates comprehensive testing for Redis and Memcached implementations using Go and Fiber framework.

## Project Structure

![Coverage](https://img.shields.io/badge/dynamic/json?color=brightgreen&label=coverage&query=coverage&url=https://api.github.com/repos/{owner}/{repo}/contents/coverage.txt)

## Test Coverage

This project maintains test coverage statistics. You can view:
- Coverage details in pull request comments
- Coverage summary in GitHub Actions job summary
- Coverage badge above

![Coverage](https://img.shields.io/badge/Coverage-0%25-red)

## Test Coverage
Test coverage is automatically calculated and updated on each push.

![Tests](https://github.com/veyselaksin/StriGO/actions/workflows/tests.yml/badge.svg)

## Test Coverage
Test coverage reports are available in GitHub Actions artifacts.
