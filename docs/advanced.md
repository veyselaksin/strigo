---
layout: page
title: Advanced Usage
nav_order: 3
---

# Advanced Usage
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Dynamic Rate Limiting

Strigo allows you to apply dynamic rate limiting rules based on request properties. Here are some common scenarios:

### Query Parameter Based Limits

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
{: .highlight }

### User Type Based Limits

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
{: .highlight }

## Best Practices
{: .text-delta }

1. **Pattern Naming**
   - Use unique and meaningful pattern names
   - Include version or feature identifiers
   {: .note }

2. **Multiple Limits**
   - Define both short-term and long-term limits
   - Consider different user tiers
   {: .important }

3. **Error Handling**
   - Handle rate limit exceeds gracefully
   - Provide meaningful error messages
   {: .warning }

4. **Monitoring**
   - Log rate limiting events
   - Set up alerts for abuse
   {: .danger }

## Performance Tips

### Redis Configuration

For optimal performance with Redis:

```go
manager := ratelimiter.NewManager(limiter.Redis, "localhost:6379")
```
{: .note }

### Memcached Configuration

For Memcached setup:

```go
manager := ratelimiter.NewManager(limiter.Memcached, "localhost:11211")
```
{: .note }

[Next: API Reference](api){: .btn .btn-purple }