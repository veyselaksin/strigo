package fiber

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/veyselaksin/strigo/middleware/ratelimiter"
	"github.com/veyselaksin/strigo/pkg/config"
	"github.com/veyselaksin/strigo/pkg/duration"
	"github.com/veyselaksin/strigo/pkg/limiter"
)

// Middleware Fiber için rate limit middleware'i
type Middleware struct {
	manager *ratelimiter.Manager
}

// New creates a new Fiber rate limit middleware
func New(manager *ratelimiter.Manager) *Middleware {
	return &Middleware{
		manager: manager,
	}
}

// getUserTypeFromToken extracts user type from JWT token
func getUserTypeFromToken(c *fiber.Ctx) string {
	// Get token from Authorization header
	token := c.Get("Authorization")
	if token == "" {
		return "anonymous"
	}

	// TODO: Implement proper JWT token parsing
	// For now just return basic user type
	return "basic"
}

// buildKey creates a unique key for rate limiting
func buildKey(path string, userType string, c *fiber.Ctx) string {
	ip := c.IP()
	return fmt.Sprintf("%s:%s:%s", userType, path, ip)
}

// Handle creates a new rate limit handler with given config
func (m *Middleware) Handle(cfg limiter.Config) fiber.Handler {
	lim, err := m.manager.GetLimiter(cfg)
	if err != nil {
		panic(err) // Veya daha iyi bir hata yönetimi
	}

	return func(c *fiber.Ctx) error {
		userType := getUserTypeFromToken(c)
		key := buildKey(c.Path(), userType, c)

		if !lim.Allow(key) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		}

		return c.Next()
	}
}

// Hazır konfigürasyonlar
func (m *Middleware) StandardAPILimit() fiber.Handler {
	return m.Handle(limiter.Config{
		Rules: []limiter.RuleConfig{
			{
				Pattern:  "user:premium:",
				Period:   duration.MINUTELY,
				Limit:    100,
				Strategy: config.TokenBucket,
			},
			{
				Pattern:  "user:basic:",
				Period:   duration.MINUTELY,
				Limit:    20,
				Strategy: config.TokenBucket,
			},
		},
		Prefix: "standard-api",
	})
}

func (m *Middleware) StrictAPILimit() fiber.Handler {
	return m.Handle(limiter.Config{
		Rules: []limiter.RuleConfig{
			{
				Pattern:  "user:premium:",
				Period:   duration.MINUTELY,
				Limit:    50,
				Strategy: config.TokenBucket,
			},
			{
				Pattern:  "user:basic:",
				Period:   duration.MINUTELY,
				Limit:    10,
				Strategy: config.TokenBucket,
			},
		},
		Prefix: "strict-api",
	})
}
