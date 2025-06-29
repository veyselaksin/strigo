---
layout: page
title: API Reference
nav_order: 3
description: "Complete API documentation for StriGO v2.0.0"
---

# API Reference

Complete reference for all StriGO v2.0.0 functions, types, and configurations.

{: .no_toc }

## Table of contents

{: .no_toc .text-delta }

1. TOC
   {:toc}

---

## Package Import

```go
import "github.com/veyselaksin/strigo/v2"
```

## Core Types

### Options

Configuration options for creating a rate limiter:

```go
type Options struct {
    Points        int64       // Maximum points that can be consumed over duration
    Duration      int64       // Time window for point consumption in seconds
    Strategy      Strategy    // Rate limiting algorithm (TokenBucket, LeakyBucket, etc.)
    BlockDuration int64       // How long to block key after limit exceeded (seconds)
    KeyPrefix     string      // Prefix used to create unique keys in storage backend
    StoreClient   interface{} // Redis/Memcached client instance (nil = memory)
    StoreType     string      // Type of store client ("redis", "memcached", "memory")
}
```

### Result

Information returned by `Consume` operations:

```go
type Result struct {
    MsBeforeNext      int64 // Milliseconds before next action can be done
    RemainingPoints   int64 // Number of remaining points in current duration
    ConsumedPoints    int64 // Number of consumed points in current duration
    IsFirstInDuration bool  // Whether the action is first in current duration
    TotalHits         int64 // Total points allowed in the duration
    Allowed           bool  // Whether the request was allowed
}
```

### Strategy

Available rate limiting strategies:

```go
type Strategy int

const (
    TokenBucket   Strategy = iota // Classic token bucket algorithm (default)
    LeakyBucket                   // Leaky bucket algorithm for smooth traffic
    FixedWindow                   // Fixed time window counting
    SlidingWindow                 // Sliding time window for accurate limiting
)
```

## Core Functions

### New

Create a new rate limiter instance:

```go
func New(opts *Options) (*RateLimiter, error)
```

**Parameters:**

- `opts`: Configuration options

**Returns:**

- `*RateLimiter`: Rate limiter instance
- `error`: Error if creation fails

**Example:**

```go
limiter, err := strigo.New(&strigo.Options{
    Points:   100,
    Duration: 60,
})
```

## RateLimiter Methods

### Consume

Consume points from the rate limiter:

```go
func (rl *RateLimiter) Consume(key string, points int64) (*Result, error)
```

**Parameters:**

- `key`: Unique identifier for the client
- `points`: Number of points to consume

**Returns:**

- `*Result`: Information about the consumption
- `error`: Error if operation fails

### Get

Get current rate limit status without consuming points:

```go
func (rl *RateLimiter) Get(key string) (*Result, error)
```

**Parameters:**

- `key`: Unique identifier for the client

**Returns:**

- `*Result`: Current status (nil if key doesn't exist)
- `error`: Error if operation fails

### Block

Manually block a key for specified duration:

```go
func (rl *RateLimiter) Block(key string, blockDurationSeconds int64) error
```

**Parameters:**

- `key`: Unique identifier for the client
- `blockDurationSeconds`: Duration to block in seconds

**Returns:**

- `error`: Error if operation fails

### Reset

Reset rate limit for a key:

```go
func (rl *RateLimiter) Reset(key string) error
```

**Parameters:**

- `key`: Unique identifier for the client

**Returns:**

- `error`: Error if operation fails

### Close

Close the rate limiter and cleanup resources:

```go
func (rl *RateLimiter) Close() error
```

**Returns:**

- `error`: Error if cleanup fails

## Result Methods

### Headers

Get standard HTTP rate limit headers:

```go
func (r *Result) Headers() map[string]string
```

**Returns:**

- `map[string]string`: HTTP headers for rate limiting

**Headers returned:**

- `X-RateLimit-Limit`: Total points allowed
- `X-RateLimit-Remaining`: Remaining points
- `X-RateLimit-Reset`: Reset time (Unix timestamp)
- `Retry-After`: Seconds to wait (if rate limited)

**Example:**

```go
result, _ := limiter.Consume("user:123", 1)
headers := result.Headers()

for name, value := range headers {
    c.Set(name, value) // Set in HTTP response
}
```

## Storage Backends

### Memory (Default)

Built-in in-memory storage:

```go
limiter, err := strigo.New(&strigo.Options{
    Points:   100,
    Duration: 60,
    // No StoreClient = memory storage
})
```

### Redis

Redis-based distributed storage:

```go
import "github.com/redis/go-redis/v9"

redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

limiter, err := strigo.New(&strigo.Options{
    Points:      100,
    Duration:    60,
    StoreClient: redisClient,
    KeyPrefix:   "myapp",
})
```

### Memcached

Memcached-based distributed storage:

```go
import "github.com/bradfitz/gomemcache/memcache"

mcClient := memcache.New("localhost:11211")

limiter, err := strigo.New(&strigo.Options{
    Points:      100,
    Duration:    60,
    StoreClient: mcClient,
    KeyPrefix:   "myapp",
})
```

## Error Handling

Common error scenarios:

```go
limiter, err := strigo.New(&strigo.Options{
    Points:   0, // Invalid: must be > 0
    Duration: 0, // Invalid: must be > 0
})
if err != nil {
    // Handle invalid configuration
}

result, err := limiter.Consume("user:123", 1)
if err != nil {
    // Handle storage errors (network, connection, etc.)
}
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/veyselaksin/strigo/v2"
)

func main() {
    // Create Redis client
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // Create rate limiter
    limiter, err := strigo.New(&strigo.Options{
        Points:        10,           // 10 requests
        Duration:      60,           // per minute
        Strategy:      strigo.TokenBucket,
        BlockDuration: 300,          // 5 minutes block
        KeyPrefix:     "api",
        StoreClient:   redisClient,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer limiter.Close()

    // Test rate limiting
    for i := 0; i < 12; i++ {
        result, err := limiter.Consume("user:123", 1)
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }

        if result.Allowed {
            fmt.Printf("Request %d: ✅ Allowed (remaining: %d)\n",
                i+1, result.RemainingPoints)
        } else {
            fmt.Printf("Request %d: ❌ Rate limited (retry in %ds)\n",
                i+1, result.MsBeforeNext/1000)
        }

        // Add headers to response
        headers := result.Headers()
        for name, value := range headers {
            fmt.Printf("  %s: %s\n", name, value)
        }

        time.Sleep(100 * time.Millisecond)
    }
}
```

[Back to Home](./){: .btn .btn-blue }

*Last synced with README.md: 2025-06-29 15:15:49 UTC*
