---
layout: home
title: Home
nav_order: 1
description: "StriGO v2.0.0 - High-performance rate limiter for Go applications"
permalink: /
---

# StriGO v2.0.0 - Production Ready Rate Limiter 🚀

[![Version](https://img.shields.io/github/v/release/veyselaksin/strigo?include_prereleases)](https://github.com/veyselaksin/strigo/releases)
[![Test Coverage](https://img.shields.io/badge/coverage-95%25-brightgreen)](https://github.com/veyselaksin/strigo/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/veyselaksin/strigo)](https://goreportcard.com/report/github.com/veyselaksin/strigo)

**StriGO** is a high-performance, production-ready rate limiter for Go applications with Redis and Memcached support.

{: .fs-6 .fw-300 }

---

## Installation

```bash
go get github.com/veyselaksin/strigo/v2@v2.0.0
```

## Quick Example

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/veyselaksin/strigo/v2"
)

func main() {
    // Create rate limiter with Redis backend
    limiter, err := strigo.NewRateLimiter(strigo.Options{
        Backend: "redis",
        RedisURL: "redis://localhost:6379",
        MemoryFallback: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer limiter.Close()

    // Rate limit: 10 requests per minute
    result, err := limiter.Consume("user:123", 1, time.Minute, 10)
    if err != nil {
        log.Fatal(err)
    }

    if result.Allowed {
        fmt.Printf("Request allowed! Remaining: %d\n", result.Remaining)
    } else {
        fmt.Printf("Rate limited! Retry after: %v\n", result.RetryAfter)
    }
}
```

### Web Framework Integration

```go
package main

import (
    "time"

    "github.com/veyselaksin/strigo/v2"
    "github.com/gofiber/fiber/v2"
)

func main() {
    // Create rate limiter
    limiter, err := strigo.NewRateLimiter(strigo.Options{
        Backend: "redis",
        RedisURL: "redis://localhost:6379",
        MemoryFallback: true,
    })
    if err != nil {
        panic(err)
    }
    defer limiter.Close()

    app := fiber.New()

    // Rate limiting middleware
    app.Use(func(c *fiber.Ctx) error {
        userID := c.Get("X-User-ID", "anonymous")

        result, err := limiter.Consume(userID, 1, time.Minute, 60)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Rate limiter error"})
        }

        if !result.Allowed {
            return c.Status(429).JSON(fiber.Map{
                "error": "Rate limit exceeded",
                "retry_after": result.RetryAfter.Seconds(),
            })
        }

        c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
        return c.Next()
    })

    app.Get("/", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"message": "Hello World!"})
    })

    app.Listen(":3000")
}
```

## Table of contents

{: .no_toc .text-delta }

1. TOC
   {:toc}

## ✨ What's New in v2.0.0

- 🚀 **Simplified API** - Clean, intuitive interface
- ⚡ **100K+ req/s Performance** - Exceptional speed
- 📊 **Professional Benchmarks** - Visual performance charts
- 🧪 **95% Test Coverage** - Comprehensive testing
- 🐳 **Docker Integration** - Complete test environments
- 📚 **Enhanced Documentation** - Better examples and guides

## Features

StriGO v2.0.0 provides comprehensive rate limiting:

- **High Performance**: 100K+ req/s concurrent throughput
- **Multiple Storage Backends**: Redis, Memcached, and in-memory
- **Point-based System**: Variable consumption per operation
- **Framework Agnostic**: Works with any Go web framework
- **Docker Ready**: Complete test environments
- **Professional Tools**: Benchmark generation and visualization
  {: .fs-6 .fw-300 }

## Quick Start

### Basic Rate Limiter

```go
package main

import (
    "fmt"
    "log"
    "github.com/veyselaksin/strigo"
)

func main() {
    // Create rate limiter - 5 requests per 10 seconds
    opts := &strigo.Options{
        Points:   5,  // 5 requests
        Duration: 10, // per 10 seconds
    }

    limiter, err := strigo.New(opts)
    if err != nil {
        log.Fatal(err)
    }
    defer limiter.Close()

    // Check rate limit
    result, err := limiter.Consume("user:123", 1)
    if err != nil {
        log.Fatal(err)
    }

    if result.Allowed {
        fmt.Printf("✅ Request allowed! Remaining: %d\n", result.RemainingPoints)
    } else {
        fmt.Printf("❌ Rate limited! Try again in %dms\n", result.MsBeforeNext)
    }
}
```

{: .highlight }

### Redis-based Rate Limiting

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/veyselaksin/strigo"
)

func main() {
    // Create Redis client
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // Create rate limiter with Redis storage
    opts := &strigo.Options{
        Points:      100, // 100 requests
        Duration:    60,  // per minute
        StoreClient: redisClient,
        KeyPrefix:   "myapp",
    }

    limiter, err := strigo.New(opts)
    if err != nil {
        log.Fatal(err)
    }
    defer limiter.Close()

    result, err := limiter.Consume("api:user456", 1)
    // Handle result...
}
```

{: .highlight }

### Variable Point Consumption

```go
// Different operations consume different amounts of points
limiter, _ := strigo.New(&strigo.Options{
    Points:   100, // 100 points total
    Duration: 60,  // per minute
})

operations := map[string]int64{
    "view_profile":    1,  // Light operation
    "update_profile":  5,  // Medium operation
    "upload_file":     10, // Heavy operation
    "generate_report": 25, // Very heavy operation
}

for operation, cost := range operations {
    result, err := limiter.Consume("user:123", cost)
    if result.Allowed {
        fmt.Printf("✅ %s allowed (cost: %d, remaining: %d)\n",
            operation, cost, result.RemainingPoints)
    } else {
        fmt.Printf("❌ %s blocked - rate limit exceeded\n", operation)
    }
}
```

{: .highlight }

## Configuration Options

### Options Structure

```go
type Options struct {
    Points        int64       // Maximum points per duration
    Duration      int64       // Time window in seconds
    KeyPrefix     string      // Prefix for storage keys
    StoreClient   interface{} // Redis/Memcached client (nil = memory)
    StoreType     string      // "redis", "memcached", "memory"
    Strategy      Strategy    // Rate limiting algorithm
    BlockDuration int64       // Block duration after limit exceeded
}
```

### Storage Backends

| Backend       | Description         | Use Case               |
| :------------ | :------------------ | :--------------------- |
| **Memory**    | Built-in storage    | Single instance apps   |
| **Redis**     | Distributed caching | Multi-instance apps    |
| **Memcached** | Fast caching        | High-performance needs |

## Performance Benchmarks

StriGO v2.0.0 delivers exceptional performance:

### Redis Performance

- **Concurrent**: 109,156 req/s ⚡️
- **Sequential**: 11,682 req/s
- **Variable Points**: 12,148 op/s

### Memcached Performance

- **Concurrent**: 89,446 req/s ⚡️
- **Sequential**: 11,773 req/s
- **Get Status**: 22,608 gets/s

_Tested on Apple M3, Go 1.22.3_

## Next Steps

- [API Reference](api) - Complete API documentation
- [Advanced Usage](advanced) - Complex scenarios and patterns
- [Docker Configuration](docker) - Container setup and testing
- [Best Practices](best-practices) - Production recommendations

[View on GitHub](https://github.com/veyselaksin/strigo){: .btn .btn-purple .mr-2 }
[API Reference](api){: .btn .btn-blue }
