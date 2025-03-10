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

StriGo provides powerful dynamic rate limiting capabilities based on various request properties. Here are some advanced usage scenarios:

### Multiple Strategy Implementation

```go
app.Get("/api/advanced", fiberMiddleware.RateLimitHandler(manager, func(c *fiber.Ctx) []strigo.RuleConfig {
    return []strigo.RuleConfig{
        {
            Pattern:  "short_term",
            Strategy: strigo.TokenBucket,
            Period:   strigo.MINUTELY,
            Limit:    100,
        },
        {
            Pattern:  "long_term",
            Strategy: strigo.SlidingWindow,
            Period:   strigo.DAILY,
            Limit:    10000,
        },
    }
}), handler)
```
{: .highlight }

### User-Based Rate Limiting

```go
app.Get("/api/user", fiberMiddleware.RateLimitHandler(manager, func(c *fiber.Ctx) []strigo.RuleConfig {
    userTier := c.Get("X-User-Tier", "free")
    
    switch userTier {
    case "premium":
        return []strigo.RuleConfig{
            {
                Pattern:  "premium_short",
                Strategy: strigo.TokenBucket,
                Period:   strigo.MINUTELY,
                Limit:    1000,
            },
            {
                Pattern:  "premium_long",
                Strategy: strigo.SlidingWindow,
                Period:   strigo.DAILY,
                Limit:    100000,
            },
        }
    case "basic":
        return []strigo.RuleConfig{
            {
                Pattern:  "basic_short",
                Strategy: strigo.TokenBucket,
                Period:   strigo.MINUTELY,
                Limit:    100,
            },
            {
                Pattern:  "basic_long",
                Strategy: strigo.LeakyBucket,
                Period:   strigo.HOURLY,
                Limit:    1000,
            },
        }
    default: // free tier
        return []strigo.RuleConfig{
            {
                Pattern:  "free_limit",
                Strategy: strigo.TokenBucket,
                Period:   strigo.MINUTELY,
                Limit:    10,
            },
        }
    }
}), handler)
```
{: .highlight }

### Resource-Based Limiting

```go
app.Post("/api/upload", fiberMiddleware.RateLimitHandler(manager, func(c *fiber.Ctx) []strigo.RuleConfig {
    contentType := c.Get("Content-Type")
    
    rules := []strigo.RuleConfig{
        {
            Pattern:  "global_upload",
            Strategy: strigo.TokenBucket,
            Period:   strigo.MINUTELY,
            Limit:    50,
        },
    }

    // Add specific limits for different content types
    switch contentType {
    case "image/jpeg", "image/png":
        rules = append(rules, strigo.RuleConfig{
            Pattern:  "image_upload",
            Strategy: strigo.LeakyBucket,
            Period:   strigo.HOURLY,
            Limit:    100,
        })
    case "video/mp4":
        rules = append(rules, strigo.RuleConfig{
            Pattern:  "video_upload",
            Strategy: strigo.SlidingWindow,
            Period:   strigo.DAILY,
            Limit:    10,
        })
    }

    return rules
}), uploadHandler)
```
{: .highlight }

## Advanced Configuration

### Custom Key Generation

```go
func customKeyGenerator(c *fiber.Ctx) string {
    // Combine IP and User-Agent for more precise rate limiting
    return fmt.Sprintf("%s:%s", c.IP(), c.Get("User-Agent"))
}

app.Use(fiberMiddleware.RateLimitHandler(manager, getRules, 
    fiberMiddleware.Config{
        KeyGenerator: customKeyGenerator,
    },
))
```
{: .highlight }

### Error Response Customization

```go
func customErrorHandler(c *fiber.Ctx) error {
    return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
        "error": "Rate limit exceeded",
        "retry_after": c.Get("X-RateLimit-Reset"),
        "limit": c.Get("X-RateLimit-Limit"),
    })
}

app.Use(fiberMiddleware.RateLimitHandler(manager, getRules, 
    fiberMiddleware.Config{
        ErrorHandler: customErrorHandler,
    },
))
```
{: .highlight }

## Best Practices
{: .text-delta }

1. **Pattern Naming**
   - Use descriptive, hierarchical patterns (e.g., `api:user:upload`)
   - Include version identifiers for API versioning
   - Keep patterns consistent across similar endpoints
   {: .note }

2. **Multiple Limits**
   - Implement both short-term and long-term limits
   - Consider resource-specific limits
   - Use different strategies for different use cases
   {: .important }

3. **Error Handling**
   - Provide clear error messages
   - Include retry-after headers
   - Log rate limit violations
   {: .warning }

4. **Performance Optimization**
   - Use appropriate storage backend for your scale
   - Implement caching where appropriate
   - Monitor storage backend performance
   {: .danger }

## Storage Configuration

### Redis Advanced Setup

```go
manager := strigo.NewManager(strigo.Redis, "localhost:6379", strigo.ManagerConfig{
    PoolSize: 100,
    MinIdleConns: 10,
    MaxRetries: 3,
    ReadTimeout: time.Second * 2,
    WriteTimeout: time.Second * 2,
})
```
{: .note }

### Memcached Advanced Setup

```go
manager := strigo.NewManager(strigo.Memcached, "localhost:11211", strigo.ManagerConfig{
    MaxIdleConns: 50,
    Timeout: time.Second * 2,
    MaxRetries: 3,
})
```
{: .note }

[Next: API Reference](api){: .btn .btn-purple }