---
layout: page
title: API Reference
nav_order: 4
---

# API Reference
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Packages

### ratelimiter

Main package providing rate limiting functionality.
{: .fs-6 .fw-300 }

#### Manager Interface

```go
type Manager interface {
    Close() error
    IsAllowed(pattern string, strategy config.Strategy, period duration.Period, limit int64) (bool, error)
}
```
{: .highlight }

#### NewManager Function

```go
func NewManager(storageType limiter.StorageType, connectionString string) Manager
```

| Parameter | Type | Description |
|:----------|:-----|:------------|
| storageType | `limiter.StorageType` | Redis or Memcached |
| connectionString | `string` | Connection details |

### middleware/fiber

Middleware implementation for the Fiber web framework.
{: .fs-6 .fw-300 }

#### RateLimitHandler

```go
func RateLimitHandler(
    manager ratelimiter.Manager,
    configProvider func(*fiber.Ctx) []limiter.RuleConfig
) fiber.Handler
```
{: .highlight }

## Error Codes

| Code | Description |
|:-----|:------------|
| 429 | Too Many Requests |
| 500 | Internal Server Error |

## Response Headers

| Header | Description |
|:-------|:------------|
| X-RateLimit-Limit | Total allowed requests |
| X-RateLimit-Remaining | Remaining requests |
| X-RateLimit-Reset | Reset timestamp |

[Back to Home](./){: .btn .btn-blue }