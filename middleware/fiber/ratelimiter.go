package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/veyselaksin/strigo"
)

// RateLimitConfig holds the configuration for rate limiting middleware
type RateLimitConfig struct {
	// RulesFunc returns rate limit rules based on the request context
	RulesFunc func(*fiber.Ctx) []strigo.LimiterConfig
	// KeyFunc generates a unique key for rate limiting (optional)
	KeyFunc func(*fiber.Ctx) string
	// ErrorHandler handles rate limit errors (optional)
	ErrorHandler fiber.Handler
}

// RateLimitHandler creates a new rate limiting middleware for Fiber
func RateLimitHandler(manager *strigo.Manager, rulesFunc func(*fiber.Ctx) []strigo.RuleConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rules := rulesFunc(c)
		if len(rules) == 0 {
			return c.Next()
		}

		// Generate key and check all rules
		for _, rule := range rules {
			key := buildKey(c, rule.Pattern)
			if !manager.Allow(key, strigo.LimiterConfig{Rules: []strigo.RuleConfig{rule}}) {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": "Rate limit exceeded",
				})
			}
		}

		return c.Next()
	}
}

func buildKey(c *fiber.Ctx, pattern string) string {
	return pattern + ":" + c.IP() + ":" + c.Path()
}
