# Advanced Strigo Usage

## Dynamic Rate Limiting
Strigo allows you to apply dynamic rate limiting rules based on request properties. Here are some common scenarios:

### 1. Query Parameter Based Limits
```go
app.Get("/api/images", fiberMiddleware.RateLimitHandler(manager, func(c *fiber.Ctx) []limiter.RuleConfig {
    queryType := c.Query("type")
    if queryType == "image" {
        return []limiter.RuleConfig{
            {
                Pattern:  "image_daily",
                Strategy: config.TokenBucket,
                Period:   duration.DAILY,
                Limit:    3,
            },
        }
    }
    return []limiter.RuleConfig{
        {
            Pattern:  "default_daily",
            Strategy: config.TokenBucket,
            Period:   duration.DAILY,
            Limit:    100,
        },
    }
}), handler)
```

### 2. Multiple Limits Based on User Type
```go
app.Get("/api/content", fiberMiddleware.RateLimitHandler(manager, func(c *fiber.Ctx) []limiter.RuleConfig {
    userType := c.Get("X-User-Type")

    switch userType {
    case "pro":
        return []limiter.RuleConfig{
            {
                Pattern:  "pro_minute",
                Strategy: config.TokenBucket,
                Period:   duration.MINUTELY,
                Limit:    100,
            },
            {
                Pattern:  "pro_daily",
                Strategy: config.TokenBucket,
                Period:   duration.DAILY,
                Limit:    10000,
            },
        }
    case "free":
        return []limiter.RuleConfig{
            {
                Pattern:  "free_minute",
                Strategy: config.TokenBucket,
                Period:   duration.MINUTELY,
                Limit:    10,
            },
            {
                Pattern:  "free_daily",
                Strategy: config.TokenBucket,
                Period:   duration.DAILY,
                Limit:    1000,
            },
        }
    default:
        return []limiter.RuleConfig{
            {
                Pattern:  "guest_minute",
                Strategy: config.TokenBucket,
                Period:   duration.MINUTELY,
                Limit:    5,
            },
        }
    }
}), handler)
```

## Using Memcached
To use Memcached instead of Redis:

```go
manager := ratelimiter.NewManager(limiter.Memcached, "localhost:11211")
```

## Best Practices
- **Pattern Naming:** Use unique and meaningful pattern names for each endpoint
- **Multiple Limits:** Define both short-term and long-term limits for critical endpoints
- **Error Handling:** Handle rate limit exceeds with appropriate HTTP status codes
- **Monitoring:** Log and monitor rate limiting events

## Performance Tips
- Use Redis cluster for scalability under high load
- Optimize the number of patterns - avoid using too many patterns per request
- Configure appropriate buffer size and connection pool settings

---
Navigation: [Home](README.md) | [Getting Started](getting-started.md) | [API Reference](api.md)