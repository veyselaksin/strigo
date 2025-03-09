---
layout: page
title: Getting Started
nav_order: 2
---

# Getting Started with Strigo
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Installation

To add Strigo to your project, run the following command:

```bash
go get github.com/veyselaksin/strigo
```
{: .highlight }

## Features

Strigo comes packed with powerful features out of the box:

- Multiple storage support (Redis, Memcached)
- Flexible rate limiting rules
- Support for Token Bucket strategy
- Define limits for different time intervals
- Integration with Fiber web framework
{: .fs-6 .fw-300 }

## Quick Start

Below is a simple example application:

```go
package main

import (
    "log"
    "github.com/gofiber/fiber/v2"
    "github.com/veyselaksin/strigo/config"
    fiberMiddleware "github.com/veyselaksin/strigo/middleware/fiber"
    "github.com/veyselaksin/strigo/middleware/ratelimiter"
    "github.com/veyselaksin/strigo/pkg/duration"
    "github.com/veyselaksin/strigo/pkg/limiter"
)

func main() {
    app := fiber.New()

    // Create rate limiter manager
    manager := ratelimiter.NewManager(limiter.Redis, "localhost:6379")
    defer manager.Close()

    // Simple rate-limited endpoint
    app.Get("/api/basic", fiberMiddleware.RateLimitHandler(manager, func(c *fiber.Ctx) []limiter.RuleConfig {
        return []limiter.RuleConfig{
            {
                Pattern:  "basic_limit",
                Strategy: config.TokenBucket,
                Period:   duration.MINUTELY,
                Limit:    10,
            },
        }
    }), func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"message": "Success"})
    })

    log.Fatal(app.Listen(":3000"))
}
```
{: .highlight }

## Configuration

### Rate Limit Rules

Each rule contains the following parameters:

| Parameter | Description |
|:----------|:------------|
| Pattern | Unique identifier for the rule |
| Strategy | Rate limiting strategy (e.g., TokenBucket) |
| Period | Time interval (MINUTELY, HOURLY, DAILY) |
| Limit | Maximum number of allowed requests |

### Storage Options

Strigo supports the following storage options:

- **Redis:** `limiter.Redis`
- **Memcached:** `limiter.Memcached`
{: .note }

[Next: Advanced Usage](advanced){: .btn .btn-purple }