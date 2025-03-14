// Package middleware provides HTTP middleware implementations for various web frameworks
// to integrate StriGO rate limiting functionality.
//
// # Overview
//
// The middleware package provides ready-to-use middleware implementations for
// popular web frameworks. These middlewares handle rate limiting at the HTTP
// request level, making it easy to protect your web applications from abuse.
//
// Supported Frameworks
//
//   - Fiber: High-performance web framework
//   - Standard net/http: Go's standard HTTP package
//   - Echo (coming soon)
//   - Gin (coming soon)
//
// Features
//
//   - Request-based rate limiting
//   - Custom key generation
//   - Response headers for rate limit information
//   - Configurable error responses
//   - Framework-specific optimizations
//
// Example Usage with Fiber
//
//	app := fiber.New()
//	config := strigo.NewDefaultConfig()
//	limiter, err := strigo.NewRateLimiter(config)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Basic usage
//	app.Use(middleware.NewFiberMiddleware(limiter))
//
//	// With custom configuration
//	app.Use(middleware.NewFiberMiddleware(limiter, middleware.Config{
//		KeyGenerator: func(c *fiber.Ctx) string {
//			return c.IP() // Rate limit by IP address
//		},
//		ErrorHandler: func(c *fiber.Ctx) error {
//			return c.Status(429).JSON(fiber.Map{
//				"error": "Too many requests",
//			})
//		},
//	}))
//
// Example Usage with net/http
//
//	config := strigo.NewDefaultConfig()
//	limiter, err := strigo.NewRateLimiter(config)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	handler := middleware.NewHTTPMiddleware(limiter)
//	http.ListenAndServe(":8080", handler(yourHandler))
//
// # Response Headers
//
// The middleware adds the following headers to responses:
//
//   - X-RateLimit-Limit: Maximum requests per period
//   - X-RateLimit-Remaining: Remaining requests in current period
//   - X-RateLimit-Reset: Time until the rate limit resets
//   - Retry-After: Seconds until next request is allowed (when limited)
//
// # Custom Key Generation
//
// You can customize how rate limit keys are generated:
//
//	middleware.Config{
//		KeyGenerator: func(c *fiber.Ctx) string {
//			return c.Get("X-API-Key") // Rate limit by API key
//		},
//	}
//
// # Error Handling
//
// Custom error handling can be configured:
//
//	middleware.Config{
//		ErrorHandler: func(c *fiber.Ctx) error {
//			return c.Status(429).JSON(fiber.Map{
//				"error": "Rate limit exceeded",
//				"retry_after": c.Get("Retry-After"),
//			})
//		},
//	}
//
// For more detailed examples and documentation, visit:
// https://veyselaksin.github.io/strigo/
package middleware
