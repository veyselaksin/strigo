---
layout: page
title: Getting Started
nav_order: 2
description: "Complete guide to setting up and using StriGO v2.0.0"
---

# Getting Started with StriGO

This guide will help you get started with StriGO v2.0.0, from installation to production deployment.

{: .no_toc }

## Table of contents

{: .no_toc .text-delta }

1. TOC
   {:toc}

---

## Installation

Install StriGO v2.0.0 using Go modules:

```bash
go get github.com/veyselaksin/strigo/v2@v2.0.0
```

Then import it in your Go code:

```go
import "github.com/veyselaksin/strigo/v2"
```

## Basic Usage

### Simple Rate Limiting

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
    "fmt"
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

## Quick Start Examples


### 📋 Prerequisites

- 🔧 Go 1.22.3 or later
- 🐳 Docker and Docker Compose
- 📦 Redis
- 💾 Memcached

### 📥 Installation

```bash
go get github.com/veyselaksin/strigo/v2@v2.0.0
```

## Web Framework Integration

### Fiber Web Framework

Create a modular structure with global limiters:

#### Step 1: Create `limiter.go`

```go
package main

import (
    "log"
    "github.com/redis/go-redis/v9"
    "github.com/veyselaksin/strigo/v2"
)

var (
    // Global rate limiter instances
    ApiLimiter    *strigo.RateLimiter
    AuthLimiter   *strigo.RateLimiter
    UploadLimiter *strigo.RateLimiter
)

func InitializeLimiters() {
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    var err error

    // API Rate Limiter - 100 requests per minute
    ApiLimiter, err = strigo.New(&strigo.Options{
        Points:      100,
        Duration:    60,
        KeyPrefix:   "api",
        StoreClient: redisClient,
    })
    if err != nil {
        log.Printf("⚠️ Redis not available for API limiter: %v", err)
        // Fallback to memory
        ApiLimiter, _ = strigo.New(&strigo.Options{
            Points:    100,
            Duration:  60,
            KeyPrefix: "api",
        })
    }

    // Auth Rate Limiter - 5 attempts per 5 minutes
    AuthLimiter, err = strigo.New(&strigo.Options{
        Points:      5,
        Duration:    300,
        KeyPrefix:   "auth",
        StoreClient: redisClient,
    })
    if err != nil {
        log.Printf("⚠️ Redis not available for auth limiter: %v", err)
        AuthLimiter, _ = strigo.New(&strigo.Options{
            Points:    5,
            Duration:  300,
            KeyPrefix: "auth",
        })
    }

    // Upload Rate Limiter - 10 uploads per hour
    UploadLimiter, err = strigo.New(&strigo.Options{
        Points:      10,
        Duration:    3600,
        KeyPrefix:   "upload",
        StoreClient: redisClient,
    })
    if err != nil {
        log.Printf("⚠️ Redis not available for upload limiter: %v", err)
        UploadLimiter, _ = strigo.New(&strigo.Options{
            Points:    10,
            Duration:  3600,
            KeyPrefix: "upload",
        })
    }

    log.Println("✅ Rate limiters initialized")
}
```

#### Step 2: Create `middleware.go`

```go
package main

import (
    "strconv"
    "time"
    "github.com/gofiber/fiber/v2"
    "github.com/veyselaksin/strigo/v2"
)

// Rate limiting middleware
func rateLimitMiddleware(limiter *strigo.RateLimiter, points ...int64) fiber.Handler {
    return func(c *fiber.Ctx) error {
        key := getUserKey(c)

        // Default to 1 point if not specified
        consumePoints := int64(1)
        if len(points) > 0 {
            consumePoints = points[0]
        }

        result, err := limiter.Consume(key, consumePoints)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{
                "error": "Rate limiter error",
            })
        }

        // Set rate limit headers
        headers := result.Headers()
        for name, value := range headers {
            c.Set(name, value)
        }

        if !result.Allowed {
            return c.Status(429).JSON(fiber.Map{
                "error":             "Rate limit exceeded",
                "retryAfterSeconds": result.MsBeforeNext / 1000,
                "retryAfterMs":      result.MsBeforeNext,
                "limit":             result.TotalHits,
                "consumed":          result.ConsumedPoints,
                "remaining":         result.RemainingPoints,
                "resetTime":         time.Now().Add(time.Duration(result.MsBeforeNext) * time.Millisecond).Unix(),
            })
        }

        c.Set("X-RateLimit-Points-Consumed", strconv.FormatInt(consumePoints, 10))
        return c.Next()
    }
}

func getUserKey(c *fiber.Ctx) string {
    // Priority: User ID > API Key > IP
    if userID := c.Get("X-User-ID"); userID != "" {
        return "user:" + userID
    }
    if apiKey := c.Get("X-API-Key"); apiKey != "" {
        return "apikey:" + apiKey
    }
    return "ip:" + c.IP()
}
```

#### Step 3: Create `main.go`

```go
package main

import (
    "log"
    "github.com/gofiber/fiber/v2"
)

func main() {
    app := fiber.New()

    // Initialize rate limiters
    InitializeLimiters()

    // API routes with different limits
    api := app.Group("/api")
    api.Get("/users", rateLimitMiddleware(ApiLimiter, 1), getUsersHandler)
    api.Get("/analytics", rateLimitMiddleware(ApiLimiter, 5), getAnalyticsHandler)  // More expensive
    api.Get("/export", rateLimitMiddleware(ApiLimiter, 10), getExportHandler)      // Very expensive

    // Authentication with stricter limits
    auth := app.Group("/auth")
    auth.Post("/login", rateLimitMiddleware(AuthLimiter, 1), loginHandler)
    auth.Post("/reset", rateLimitMiddleware(AuthLimiter, 2), resetHandler)

    // File uploads
    app.Post("/upload", rateLimitMiddleware(UploadLimiter, 1), uploadHandler)

    log.Println("🚀 Server starting on :3000")
    log.Fatal(app.Listen(":3000"))
}

// Handler functions
func getUsersHandler(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"users": []string{"user1", "user2"}})
}

func getAnalyticsHandler(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"analytics": "expensive data processing"})
}

func getExportHandler(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"export": "very expensive operation"})
}

func loginHandler(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"token": "jwt-token-here"})
}

func resetHandler(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"message": "password reset email sent"})
}

func uploadHandler(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"uploaded": true, "size": "2.3MB"})
}
```

## Advanced Usage Patterns

### 1. Multiple Rate Limiters

```go
// Different limiters for different use cases
apiLimiter, _ := strigo.New(&strigo.Options{Points: 1000, Duration: 3600})  // 1000/hour
authLimiter, _ := strigo.New(&strigo.Options{Points: 5, Duration: 300})     // 5 attempts/5min
uploadLimiter, _ := strigo.New(&strigo.Options{Points: 10, Duration: 3600}) // 10 uploads/hour
```

### 2. Check Status Without Consuming

```go
// Check current status without consuming points
result, err := limiter.Get("user:123")
if result != nil {
    fmt.Printf("Usage: %d/%d points used\n",
        result.ConsumedPoints, result.TotalHits)
}
```

### 3. Manual Blocking

```go
// Block a user for 5 minutes (300 seconds)
err := limiter.Block("user:spam123", 300)
if err != nil {
    log.Printf("Failed to block user: %v", err)
}
```

### 4. Reset Rate Limits

```go
// Reset rate limit for a key (admin operation)
err := limiter.Reset("user:123")
if err != nil {
    log.Printf("Failed to reset limit: %v", err)
}
```

## Testing Your Rate Limiter

### Basic Test Script

```bash
#!/bin/bash
# test-rate-limit.sh

echo "Testing rate limiter..."

for i in {1..8}; do
    echo "Request $i:"
    curl -s "http://localhost:3000/api/users" | jq -r '.error // "Success"'
    sleep 1
done
```

### Load Testing

```bash
# Install hey for load testing
go install github.com/rakyll/hey@latest

# Test with 100 requests, 10 concurrent
hey -n 100 -c 10 http://localhost:3000/api/users
```

## Error Handling Best Practices

```go
func handleRateLimit(limiter *strigo.RateLimiter, key string, points int64) error {
    result, err := limiter.Consume(key, points)
    if err != nil {
        // Log storage errors but don't block requests
        log.Printf("Rate limiter error: %v", err)
        return nil // Allow request to proceed
    }

    if !result.Allowed {
        return fmt.Errorf("rate limit exceeded, retry in %d seconds",
            result.MsBeforeNext/1000)
    }

    return nil
}
```

## Configuration Tips

### Production Configuration

```go
// Production settings
opts := &strigo.Options{
    Points:        1000,         // Higher limits for production
    Duration:      3600,         // 1 hour window
    KeyPrefix:     "prod-api",   // Environment-specific prefix
    BlockDuration: 900,          // 15 minutes block
    Strategy:      strigo.TokenBucket,
    StoreClient:   redisClient,  // Always use Redis in production
}
```

### Development Configuration

```go
// Development settings
opts := &strigo.Options{
    Points:   10000,     // Very high limits for testing
    Duration: 60,        // Short windows for quick testing
    KeyPrefix: "dev-api",
    // No StoreClient = memory storage for simplicity
}
```

## Next Steps

- [API Reference](api) - Complete API documentation
- [Advanced Usage](advanced) - Complex patterns and strategies
- [Best Practices](best-practices) - Production recommendations
- [Docker Setup](docker) - Container configuration

## Performance Notes

StriGO v2.0.0 delivers exceptional performance:

- **Memory**: No network overhead, instant operations
- **Redis**: 109K+ req/s concurrent performance
- **Memcached**: 89K+ req/s concurrent performance

Choose your storage backend based on your scaling needs:

- **Memory**: Single instance applications
- **Redis**: Multi-instance with persistence needs
- **Memcached**: Multi-instance with maximum performance

[Next: API Reference](api){: .btn .btn-purple }

*Last synced with README.md: 2025-07-09 22:08:55 UTC
