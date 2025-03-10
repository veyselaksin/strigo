# Getting Started with StriGo

## Installation

To add StriGo to your project, run:

```bash
go get github.com/veyselaksin/strigo
```

## Basic Usage

### Standalone Usage

Here's a simple example showing how to use StriGo directly:

```go
package main

import (
    "log"
    "github.com/veyselaksin/strigo"
)

func main() {
    // Create a new rate limiter with Redis backend
    limiter, err := strigo.NewLimiter(strigo.LimiterConfig{
        Backend: strigo.Redis,
        Address: "localhost:6379",
        Rules: []strigo.RuleConfig{
            {
                Pattern:  "api_requests",
                Strategy: strigo.TokenBucket,
                Period:   strigo.MINUTELY,
                Limit:    100,
            },
        },
        Prefix: "myapp",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer limiter.Close()

    // Check if request is allowed
    if limiter.Allow("user123") {
        // Handle request
    } else {
        // Rate limit exceeded
    }
}
```

### Web Framework Integration

StriGo integrates seamlessly with the Fiber web framework:

```go
package main

import (
    "log"
    "github.com/gofiber/fiber/v2"
    "github.com/veyselaksin/strigo"
    fiberMiddleware "github.com/veyselaksin/strigo/middleware/fiber"
)

func main() {
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

    app.Get("/api", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"message": "Success"})
    })

    log.Fatal(app.Listen(":3000"))
}
```

## Storage Backends

### Redis Configuration

```go
limiter, err := strigo.NewLimiter(strigo.LimiterConfig{
    Backend: strigo.Redis,
    Address: "localhost:6379",
    Rules: []strigo.RuleConfig{
        {
            Pattern:  "api_requests",
            Strategy: strigo.TokenBucket,
            Period:   strigo.MINUTELY,
            Limit:    100,
        },
    },
})
```

### Memcached Configuration

```go
limiter, err := strigo.NewLimiter(strigo.LimiterConfig{
    Backend: strigo.Memcached,
    Address: "localhost:11211",
    Rules: []strigo.RuleConfig{
        {
            Pattern:  "api_requests",
            Strategy: strigo.TokenBucket,
            Period:   strigo.MINUTELY,
            Limit:    100,
        },
    },
})
```

## Rate Limiting Strategies

StriGo supports multiple rate limiting strategies:

- **Token Bucket (Default)**: Simple counter-based rate limiting
- **Leaky Bucket**: Smooths out request processing
- **Fixed Window**: Resets counter at fixed intervals
- **Sliding Window**: More accurate rate limiting over time

Example with different strategies:

```go
rules := []strigo.RuleConfig{
    {
        Pattern:  "token_bucket_rule",
        Strategy: strigo.TokenBucket,
        Period:   strigo.MINUTELY,
        Limit:    100,
    },
    {
        Pattern:  "sliding_window_rule",
        Strategy: strigo.SlidingWindow,
        Period:   strigo.HOURLY,
        Limit:    1000,
    },
}
```

## Time Windows

Available time windows for rate limiting:

- `SECONDLY`: Per second
- `MINUTELY`: Per minute
- `HOURLY`: Per hour
- `DAILY`: Per day
- `WEEKLY`: Per week
- `MONTHLY`: Per month (30 days)
- `YEARLY`: Per year (365 days)

Example with multiple time windows:

```go
rules := []strigo.RuleConfig{
    {
        Pattern:  "short_term",
        Strategy: strigo.TokenBucket,
        Period:   strigo.MINUTELY,
        Limit:    100,
    },
    {
        Pattern:  "long_term",
        Strategy: strigo.TokenBucket,
        Period:   strigo.DAILY,
        Limit:    10000,
    },
}
```

## Error Handling

Always handle errors appropriately:

```go
limiter, err := strigo.NewLimiter(config)
if err != nil {
    log.Fatalf("Failed to create rate limiter: %v", err)
}
defer limiter.Close()

// Check rate limit
if !limiter.Allow("user123") {
    // Handle rate limit exceeded
    return errors.New("rate limit exceeded")
}
```

## Next Steps

- Check out the [Advanced Usage Guide](advanced.md) for more complex scenarios
- See the [API Reference](api.md) for detailed documentation
- Visit our [GitHub repository](https://github.com/veyselaksin/strigo) for the latest updates