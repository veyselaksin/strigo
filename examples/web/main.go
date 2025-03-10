package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/veyselaksin/strigo"
	fiberMiddleware "github.com/veyselaksin/strigo/middleware/fiber"
)

func main() {
	app := fiber.New()

	// Create Redis manager
	redisManager := strigo.NewManager(strigo.Redis, "localhost:6379")
	defer redisManager.Close()

	// Create Memcached manager
	memcachedManager := strigo.NewManager(strigo.Memcached, "localhost:11211")
	defer memcachedManager.Close()

	// Redis-protected endpoint
	app.Get("/redis", fiberMiddleware.RateLimitHandler(redisManager, func(c *fiber.Ctx) []strigo.RuleConfig {
		return []strigo.RuleConfig{
			{
				Pattern:  "redis_api",
				Strategy: strigo.TokenBucket,
				Period:   strigo.MINUTELY,
				Limit:    5,
			},
		}
	}), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello from Redis-protected endpoint!",
		})
	})

	// Memcached-protected endpoint
	app.Get("/memcached", fiberMiddleware.RateLimitHandler(memcachedManager, func(c *fiber.Ctx) []strigo.RuleConfig {
		return []strigo.RuleConfig{
			{
				Pattern:  "memcached_api",
				Strategy: strigo.TokenBucket,
				Period:   strigo.MINUTELY,
				Limit:    5,
			},
		}
	}), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello from Memcached-protected endpoint!",
		})
	})

	// Advanced example with multiple rules and different strategies
	app.Get("/advanced", fiberMiddleware.RateLimitHandler(redisManager, func(c *fiber.Ctx) []strigo.RuleConfig {
		return []strigo.RuleConfig{
			{
				Pattern:  "minutely_limit",
				Strategy: strigo.TokenBucket,
				Period:   strigo.MINUTELY,
				Limit:    10,
			},
			{
				Pattern:  "hourly_limit",
				Strategy: strigo.SlidingWindow,
				Period:   strigo.HOURLY,
				Limit:    100,
			},
		}
	}), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello from advanced rate-limited endpoint!",
		})
	})

	log.Fatal(app.Listen(":3000"))
}
