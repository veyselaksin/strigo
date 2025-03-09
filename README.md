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
- Send an [Email](mailto:veyselaksin@gmail.com)

---

Made with ❤️ by [Veysel Aksin](https://github.com/veyselaksin)
