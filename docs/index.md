---
layout: page
title: Getting Started
nav_order: 2
---

# Getting Started with StriGo
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Installation

To add StriGo to your project, run:

```bash
go get github.com/veyselaksin/strigo
```
{: .highlight }

## Features

StriGo provides powerful rate limiting capabilities:

- **Multiple Storage Backends**: Redis and Memcached support
- **Flexible Rate Limiting**: Multiple strategies and time windows
- **Advanced Strategies**: Token Bucket, Leaky Bucket, Sliding Window
- **Framework Integration**: Built-in Fiber middleware
- **Dynamic Configuration**: Runtime rule updates
- **High Performance**: Optimized for scale
{: .fs-6 .fw-300 }

## Quick Start

### Basic Rate Limiter

```go
package main

import (
    "log"
    "github.com/veyselaksin/strigo"
)

func main() {
    // Create a new rate limiter
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

    // Use the rate limiter
    if limiter.Allow("user123") {
        // Handle request
    } else {
        // Rate limit exceeded
    }
}
```
{: .highlight }

### Web Framework Integration

```go
package main

import (
    "log"
    "github.com/gofiber/fiber/v2"
    "github.com/veyselaksin/strigo"
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
            {
                Pattern:  "daily_limit",
                Strategy: strigo.SlidingWindow,
                Period:   strigo.DAILY,
                Limit:    1000,
            },
        }
    }))

    // Define your routes
    app.Get("/api", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"message": "Success"})
    })

    log.Fatal(app.Listen(":3000"))
}
```
{: .highlight }

## Configuration Options

### Rate Limit Rules

Each rule defines how requests are limited:

| Parameter | Description | Example |
|:----------|:------------|:--------|
| Pattern | Rule identifier | `"api_requests"` |
| Strategy | Limiting algorithm | `strigo.TokenBucket` |
| Period | Time window | `strigo.MINUTELY` |
| Limit | Request limit | `100` |

### Available Strategies

StriGo supports multiple rate limiting strategies:

- **Token Bucket**: Simple, efficient rate limiting
- **Leaky Bucket**: Smooth request processing
- **Fixed Window**: Reset-based limiting
- **Sliding Window**: Accurate, rolling window limiting

### Time Windows

Configure limits for different time periods:

```go
rules := []strigo.RuleConfig{
    {
        Pattern:  "minutely",
        Strategy: strigo.TokenBucket,
        Period:   strigo.MINUTELY,
        Limit:    100,
    },
    {
        Pattern:  "hourly",
        Strategy: strigo.SlidingWindow,
        Period:   strigo.HOURLY,
        Limit:    1000,
    },
    {
        Pattern:  "daily",
        Strategy: strigo.LeakyBucket,
        Period:   strigo.DAILY,
        Limit:    10000,
    },
}
```
{: .highlight }

## Storage Options

### Redis Setup

```go
limiter, err := strigo.NewLimiter(strigo.LimiterConfig{
    Backend: strigo.Redis,
    Address: "localhost:6379",
    Rules:   rules,
})
```
{: .note }

### Memcached Setup

```go
limiter, err := strigo.NewLimiter(strigo.LimiterConfig{
    Backend: strigo.Memcached,
    Address: "localhost:11211",
    Rules:   rules,
})
```
{: .note }

[Next: Advanced Usage](advanced){: .btn .btn-purple }