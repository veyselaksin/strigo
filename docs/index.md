---
layout: page
title: Getting Started
nav_order: 2
---

# StriGO v2.0.0 Documentation

{: .no_toc }

[![Version](https://img.shields.io/github/v/release/veyselaksin/strigo?include_prereleases)](https://github.com/veyselaksin/strigo/releases)
[![Test Coverage](https://img.shields.io/badge/coverage-95%25-brightgreen)](https://github.com/veyselaksin/strigo/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/veyselaksin/strigo)](https://goreportcard.com/report/github.com/veyselaksin/strigo)

## Table of contents

{: .no_toc .text-delta }

1. TOC
   {:toc}

## Installation

Add StriGO v2.0.0 to your project:

```bash
go get github.com/veyselaksin/strigo@v2.0.0
```

{: .highlight }

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
