---
layout: page
title: Advanced Usage
nav_order: 4
description: "Advanced patterns and production strategies for StriGO v2.0.0"
---

# Advanced Usage

{: .no_toc }

Advanced patterns and strategies for StriGO v2.0.0

## Table of contents

{: .no_toc .text-delta }

1. TOC
   {:toc}

## Production Architecture Patterns

### Global Rate Limiter Management

Create a centralized rate limiter manager for your application:

```go
// pkg/ratelimit/manager.go
package ratelimit

import (
    "log"
    "sync"
    "github.com/redis/go-redis/v9"
    "github.com/veyselaksin/strigo/v2"
)

type Manager struct {
    limiters map[string]*strigo.RateLimiter
    client   *redis.Client
    mu       sync.RWMutex
}

func NewManager(redisAddr string) *Manager {
    client := redis.NewClient(&redis.Options{
        Addr: redisAddr,
    })

    return &Manager{
        limiters: make(map[string]*strigo.RateLimiter),
        client:   client,
    }
}

func (m *Manager) GetLimiter(name string, opts *strigo.Options) *strigo.RateLimiter {
    m.mu.RLock()
    if limiter, exists := m.limiters[name]; exists {
        m.mu.RUnlock()
        return limiter
    }
    m.mu.RUnlock()

    m.mu.Lock()
    defer m.mu.Unlock()

    // Double-check after acquiring write lock
    if limiter, exists := m.limiters[name]; exists {
        return limiter
    }

    // Set Redis client if not provided
    if opts.StoreClient == nil {
        opts.StoreClient = m.client
    }

    limiter, err := strigo.New(opts)
    if err != nil {
        log.Printf("Failed to create limiter %s: %v", name, err)
        // Return memory-based fallback
        fallbackOpts := *opts
        fallbackOpts.StoreClient = nil
        limiter, _ = strigo.New(&fallbackOpts)
    }

    m.limiters[name] = limiter
    return limiter
}

func (m *Manager) Close() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    for _, limiter := range m.limiters {
        limiter.Close()
    }
    return m.client.Close()
}
```

{: .highlight }

### Tier-Based Rate Limiting

Implement different limits based on user tiers:

```go
// pkg/ratelimit/tiers.go
package ratelimit

import "github.com/veyselaksin/strigo/v2"

type UserTier string

const (
    TierFree     UserTier = "free"
    TierBasic    UserTier = "basic"
    TierPremium  UserTier = "premium"
    TierEnterprise UserTier = "enterprise"
)

type TierConfig struct {
    APIPoints    int64
    APIDuration  int64
    FilePoints   int64
    FileDuration int64
}

var TierConfigs = map[UserTier]TierConfig{
    TierFree: {
        APIPoints:    100,   // 100 requests
        APIDuration:  3600,  // per hour
        FilePoints:   5,     // 5 uploads
        FileDuration: 3600,  // per hour
    },
    TierBasic: {
        APIPoints:    1000,  // 1000 requests
        APIDuration:  3600,  // per hour
        FilePoints:   50,    // 50 uploads
        FileDuration: 3600,  // per hour
    },
    TierPremium: {
        APIPoints:    10000, // 10K requests
        APIDuration:  3600,  // per hour
        FilePoints:   500,   // 500 uploads
        FileDuration: 3600,  // per hour
    },
    TierEnterprise: {
        APIPoints:    100000, // 100K requests
        APIDuration:  3600,   // per hour
        FilePoints:   5000,   // 5K uploads
        FileDuration: 3600,   // per hour
    },
}

func (m *Manager) GetAPILimiter(tier UserTier) *strigo.RateLimiter {
    config := TierConfigs[tier]
    return m.GetLimiter(string(tier)+"_api", &strigo.Options{
        Points:    config.APIPoints,
        Duration:  config.APIDuration,
        KeyPrefix: "api_" + string(tier),
    })
}

func (m *Manager) GetFileLimiter(tier UserTier) *strigo.RateLimiter {
    config := TierConfigs[tier]
    return m.GetLimiter(string(tier)+"_file", &strigo.Options{
        Points:    config.FilePoints,
        Duration:  config.FileDuration,
        KeyPrefix: "file_" + string(tier),
    })
}
```

{: .highlight }

### Smart Middleware with User Detection

```go
// middleware/ratelimit.go
package middleware

import (
    "log"
    "strconv"
    "strings"
    "github.com/gofiber/fiber/v2"
    "github.com/veyselaksin/strigo/v2"
    "yourapp/pkg/ratelimit"
)

type RateLimitConfig struct {
    Manager       *ratelimit.Manager
    GetUserTier   func(*fiber.Ctx) ratelimit.UserTier
    GetUserKey    func(*fiber.Ctx) string
    OperationCost int64
    LimiterType   string // "api" or "file"
}

func RateLimit(config RateLimitConfig) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Get user information
        userTier := config.GetUserTier(c)
        userKey := config.GetUserKey(c)

        // Get appropriate limiter
        var limiter *strigo.RateLimiter
        switch config.LimiterType {
        case "file":
            limiter = config.Manager.GetFileLimiter(userTier)
        default:
            limiter = config.Manager.GetAPILimiter(userTier)
        }

        // Determine cost
        cost := config.OperationCost
        if cost == 0 {
            cost = 1
        }

        // Check rate limit
        result, err := limiter.Consume(userKey, cost)
        if err != nil {
            // Log error but don't block request
            log.Printf("Rate limiter error: %v", err)
            return c.Next()
        }

        // Set headers
        headers := result.Headers()
        for name, value := range headers {
            c.Set(name, value)
        }

        // Set additional headers
        c.Set("X-User-Tier", string(userTier))
        c.Set("X-Operation-Cost", strconv.FormatInt(cost, 10))

        if !result.Allowed {
            return c.Status(429).JSON(fiber.Map{
                "error":      "Rate limit exceeded",
                "user_tier":  string(userTier),
                "limit":      result.TotalHits,
                "consumed":   result.ConsumedPoints,
                "remaining":  result.RemainingPoints,
                "reset_in":   result.MsBeforeNext / 1000,
                "retry_after": result.MsBeforeNext / 1000,
            })
        }

        return c.Next()
    }
}

// Helper functions
func GetUserTierFromJWT(c *fiber.Ctx) ratelimit.UserTier {
    auth := c.Get("Authorization")
    if !strings.HasPrefix(auth, "Bearer ") {
        return ratelimit.TierFree
    }

    // Parse JWT and extract tier (simplified)
    token := auth[7:]
    if tier := parseJWTTier(token); tier != "" {
        return ratelimit.UserTier(tier)
    }

    return ratelimit.TierFree
}

func GetUserKeyFromRequest(c *fiber.Ctx) string {
    // Priority: User ID from JWT > API Key > IP
    if userID := getUserIDFromJWT(c); userID != "" {
        return "user:" + userID
    }

    if apiKey := c.Get("X-API-Key"); apiKey != "" {
        return "apikey:" + apiKey
    }

    return "ip:" + c.IP()
}
```

{: .highlight }

## Advanced Patterns

### Circuit Breaker Pattern

Combine rate limiting with circuit breaker for resilience:

```go
package circuitbreaker

import (
    "context"
    "errors"
    "sync"
    "time"
    "github.com/veyselaksin/strigo/v2"
)

type CircuitBreaker struct {
    limiter    *strigo.RateLimiter
    mu         sync.RWMutex
    state      State
    failures   int64
    lastFail   time.Time
    threshold  int64
    timeout    time.Duration
}

type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

func NewCircuitBreaker(limiter *strigo.RateLimiter, threshold int64, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        limiter:   limiter,
        state:     StateClosed,
        threshold: threshold,
        timeout:   timeout,
    }
}

func (cb *CircuitBreaker) Call(ctx context.Context, key string, points int64, fn func() error) error {
    if !cb.canProceed() {
        return errors.New("circuit breaker is open")
    }

    // Check rate limit
    result, err := cb.limiter.Consume(key, points)
    if err != nil {
        cb.recordFailure()
        return err
    }

    if !result.Allowed {
        cb.recordFailure()
        return errors.New("rate limit exceeded")
    }

    // Execute function
    err = fn()
    if err != nil {
        cb.recordFailure()
        return err
    }

    cb.recordSuccess()
    return nil
}

func (cb *CircuitBreaker) canProceed() bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()

    switch cb.state {
    case StateClosed:
        return true
    case StateOpen:
        return time.Since(cb.lastFail) > cb.timeout
    case StateHalfOpen:
        return true
    default:
        return false
    }
}

func (cb *CircuitBreaker) recordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    cb.failures = 0
    cb.state = StateClosed
}

func (cb *CircuitBreaker) recordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    cb.failures++
    cb.lastFail = time.Now()

    if cb.failures >= cb.threshold {
        cb.state = StateOpen
    }
}
```

{: .highlight }

### Background Rate Limit Monitoring

Monitor rate limit usage in real-time:

```go
package monitoring

import (
    "context"
    "log"
    "time"
    "github.com/veyselaksin/strigo/v2"
)

type Monitor struct {
    limiter *strigo.RateLimiter
    keys    []string
    interval time.Duration
    alerts   chan Alert
}

type Alert struct {
    Key        string
    Usage      float64
    Threshold  float64
    Timestamp  time.Time
}

func NewMonitor(limiter *strigo.RateLimiter, keys []string, interval time.Duration) *Monitor {
    return &Monitor{
        limiter:  limiter,
        keys:     keys,
        interval: interval,
        alerts:   make(chan Alert, 100),
    }
}

func (m *Monitor) Start(ctx context.Context) {
    ticker := time.NewTicker(m.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.checkUsage()
        }
    }
}

func (m *Monitor) checkUsage() {
    for _, key := range m.keys {
        result, err := m.limiter.Get(key)
        if err != nil || result == nil {
            continue
        }

        usage := float64(result.ConsumedPoints) / float64(result.TotalHits)

        // Alert if usage > 80%
        if usage > 0.8 {
            select {
            case m.alerts <- Alert{
                Key:       key,
                Usage:     usage,
                Threshold: 0.8,
                Timestamp: time.Now(),
            }:
            default:
                // Channel full, skip
            }
        }
    }
}

func (m *Monitor) Alerts() <-chan Alert {
    return m.alerts
}

// Usage example
func StartMonitoring(limiter *strigo.RateLimiter) {
    monitor := NewMonitor(limiter, []string{
        "user:123", "user:456", "api:premium",
    }, time.Minute)

    ctx := context.Background()
    go monitor.Start(ctx)

    go func() {
        for alert := range monitor.Alerts() {
            log.Printf("⚠️ Rate limit alert: %s at %.1f%% usage",
                alert.Key, alert.Usage*100)
        }
    }()
}
```

{: .highlight }

## Performance Optimization

### Connection Pooling

Optimize Redis connections for high-throughput applications:

```go
package config

import (
    "time"
    "github.com/redis/go-redis/v9"
)

func CreateOptimizedRedisClient(addr string) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr: addr,

        // Connection pool settings
        PoolSize:        20,              // Max number of socket connections
        MinIdleConns:    5,               // Minimum idle connections
        MaxIdleConns:    10,              // Maximum idle connections
        PoolTimeout:     4 * time.Second, // Pool timeout
        IdleTimeout:     5 * time.Minute, // Idle connection timeout

        // Command timeouts
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,

        // Retry settings
        MaxRetries:      3,
        MinRetryBackoff: 8 * time.Millisecond,
        MaxRetryBackoff: 512 * time.Millisecond,
    })
}
```

{: .highlight }

### Batch Operations

For high-volume scenarios, implement batch checking:

```go
package batch

import (
    "context"
    "sync"
    "github.com/veyselaksin/strigo/v2"
)

type BatchProcessor struct {
    limiter   *strigo.RateLimiter
    batchSize int
    workers   int
}

type Request struct {
    Key    string
    Points int64
    Result chan *strigo.Result
    Error  chan error
}

func NewBatchProcessor(limiter *strigo.RateLimiter, batchSize, workers int) *BatchProcessor {
    return &BatchProcessor{
        limiter:   limiter,
        batchSize: batchSize,
        workers:   workers,
    }
}

func (bp *BatchProcessor) ProcessBatch(ctx context.Context, requests []Request) {
    jobs := make(chan Request, len(requests))
    var wg sync.WaitGroup

    // Start workers
    for i := 0; i < bp.workers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for req := range jobs {
                result, err := bp.limiter.Consume(req.Key, req.Points)
                if err != nil {
                    req.Error <- err
                } else {
                    req.Result <- result
                }
            }
        }()
    }

    // Send jobs
    go func() {
        defer close(jobs)
        for _, req := range requests {
            select {
            case jobs <- req:
            case <-ctx.Done():
                return
            }
        }
    }()

    wg.Wait()
}
```

{: .highlight }

## Error Handling & Resilience

### Graceful Degradation

Handle storage failures gracefully:

```go
package resilience

import (
    "log"
    "sync"
    "time"
    "github.com/veyselaksin/strigo/v2"
)

type ResilientLimiter struct {
    primary   *strigo.RateLimiter
    fallback  *strigo.RateLimiter
    mu        sync.RWMutex
    healthy   bool
    lastCheck time.Time
    checkInterval time.Duration
}

func NewResilientLimiter(primary, fallback *strigo.RateLimiter) *ResilientLimiter {
    return &ResilientLimiter{
        primary:   primary,
        fallback:  fallback,
        healthy:   true,
        checkInterval: time.Minute,
    }
}

func (rl *ResilientLimiter) Consume(key string, points int64) (*strigo.Result, error) {
    limiter := rl.getCurrentLimiter()

    result, err := limiter.Consume(key, points)
    if err != nil && rl.isPrimary(limiter) {
        rl.markUnhealthy()
        // Retry with fallback
        return rl.fallback.Consume(key, points)
    }

    return result, err
}

func (rl *ResilientLimiter) getCurrentLimiter() *strigo.RateLimiter {
    rl.mu.RLock()
    defer rl.mu.RUnlock()

    if rl.healthy {
        return rl.primary
    }

    // Check if it's time to test primary again
    if time.Since(rl.lastCheck) > rl.checkInterval {
        go rl.healthCheck()
    }

    return rl.fallback
}

func (rl *ResilientLimiter) healthCheck() {
    // Simple health check - try to get status of a test key
    _, err := rl.primary.Get("health:check")

    rl.mu.Lock()
    defer rl.mu.Unlock()

    rl.lastCheck = time.Now()
    if err == nil {
        rl.healthy = true
        log.Println("✅ Primary rate limiter is healthy again")
    }
}

func (rl *ResilientLimiter) markUnhealthy() {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    if rl.healthy {
        rl.healthy = false
        log.Println("⚠️ Primary rate limiter is unhealthy, switching to fallback")
    }
}

func (rl *ResilientLimiter) isPrimary(limiter *strigo.RateLimiter) bool {
    return limiter == rl.primary
}
```

{: .highlight }

## Testing Strategies

### Load Testing Helper

```go
package testing

import (
    "context"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
    "github.com/veyselaksin/strigo/v2"
)

type LoadTestConfig struct {
    Limiter     *strigo.RateLimiter
    Concurrent  int
    Duration    time.Duration
    KeyPrefix   string
    PointsPerOp int64
}

type LoadTestResult struct {
    TotalRequests   int64
    AllowedRequests int64
    BlockedRequests int64
    ErrorRequests   int64
    Throughput      float64
    Duration        time.Duration
}

func RunLoadTest(config LoadTestConfig) *LoadTestResult {
    var (
        totalReqs   int64
        allowedReqs int64
        blockedReqs int64
        errorReqs   int64
    )

    ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
    defer cancel()

    start := time.Now()
    var wg sync.WaitGroup

    for i := 0; i < config.Concurrent; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()

            for {
                select {
                case <-ctx.Done():
                    return
                default:
                    key := fmt.Sprintf("%s:worker:%d", config.KeyPrefix, workerID)
                    result, err := config.Limiter.Consume(key, config.PointsPerOp)

                    atomic.AddInt64(&totalReqs, 1)

                    if err != nil {
                        atomic.AddInt64(&errorReqs, 1)
                    } else if result.Allowed {
                        atomic.AddInt64(&allowedReqs, 1)
                    } else {
                        atomic.AddInt64(&blockedReqs, 1)
                    }
                }
            }
        }(i)
    }

    wg.Wait()
    duration := time.Since(start)

    return &LoadTestResult{
        TotalRequests:   totalReqs,
        AllowedRequests: allowedReqs,
        BlockedRequests: blockedReqs,
        ErrorRequests:   errorReqs,
        Throughput:      float64(totalReqs) / duration.Seconds(),
        Duration:        duration,
    }
}
```

{: .highlight }

## Security Considerations

### IP-based Protection

```go
package security

import (
    "log"
    "net"
    "strings"
    "github.com/gofiber/fiber/v2"
    "github.com/veyselaksin/strigo/v2"
)

type SecurityConfig struct {
    IPLimiter    *strigo.RateLimiter
    WhitelistedIPs []string
    Strictness   int64 // Points to consume for suspicious activity
}

func SecurityMiddleware(config SecurityConfig) fiber.Handler {
    whitelist := make(map[string]bool)
    for _, ip := range config.WhitelistedIPs {
        whitelist[ip] = true
    }

    return func(c *fiber.Ctx) error {
        clientIP := c.IP()

        // Skip whitelisted IPs
        if whitelist[clientIP] {
            return c.Next()
        }

        // Determine suspicion level
        suspicionPoints := config.Strictness

        // Check for suspicious patterns
        if isSuspiciousRequest(c) {
            suspicionPoints *= 3
        }

        // Apply rate limiting
        result, err := config.IPLimiter.Consume("ip:"+clientIP, suspicionPoints)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Security check failed"})
        }

        if !result.Allowed {
            // Log security event
            log.Printf("🚨 Security: IP %s blocked (suspicious activity)", clientIP)

            return c.Status(429).JSON(fiber.Map{
                "error": "Too many requests from this IP",
                "retry_after": result.MsBeforeNext / 1000,
            })
        }

        return c.Next()
    }
}

func isSuspiciousRequest(c *fiber.Ctx) bool {
    userAgent := c.Get("User-Agent")
    path := c.Path()

    // Bot detection
    botPatterns := []string{"bot", "crawler", "spider", "scraper"}
    userAgentLower := strings.ToLower(userAgent)
    for _, pattern := range botPatterns {
        if strings.Contains(userAgentLower, pattern) {
            return true
        }
    }

    // Path traversal attempts
    if strings.Contains(path, "..") || strings.Contains(path, "//") {
        return true
    }

    // SQL injection patterns
    sqlPatterns := []string{"union", "select", "drop", "insert", "'", "\""}
    pathLower := strings.ToLower(path)
    for _, pattern := range sqlPatterns {
        if strings.Contains(pathLower, pattern) {
            return true
        }
    }

    return false
}
```

{: .highlight }

## Best Practices Summary

### 1. Architecture

- Use centralized rate limiter management
- Implement graceful degradation with fallback storage
- Design for horizontal scaling with Redis
  {: .note }

### 2. Performance

- Pool Redis connections for high throughput
- Use batch processing for bulk operations
- Monitor and alert on rate limit usage
  {: .important }

### 3. Security

- Implement IP-based rate limiting for DDoS protection
- Use different limits for different user tiers
- Log and monitor suspicious activity
  {: .warning }

### 4. Operations

- Test rate limiters under load
- Have monitoring and alerting in place
- Plan for storage backend failures
  {: .danger }

[Back to Home](./){: .btn .btn-blue .mr-2 }
[Next: API Reference](api){: .btn .btn-purple }
