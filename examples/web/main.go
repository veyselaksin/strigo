package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/veyselaksin/strigo/v2"
)

func main() {
	app := fiber.New()

	// Create rate limiters for different endpoints
	
	// API rate limiter - 100 requests per minute
	apiLimiter, err := strigo.New(&strigo.Options{
		Points:      100,
		Duration:    60,
		KeyPrefix:   "api",
		StoreClient: createRedisClient(), // Use Redis for production
	})
	if err != nil {
		log.Printf("Failed to create API limiter, using memory: %v", err)
		// Fallback to memory if Redis is not available
		apiLimiter, _ = strigo.New(&strigo.Options{
			Points:    100,
			Duration:  60,
			KeyPrefix: "api",
		})
	}
	defer apiLimiter.Close()

	// Login rate limiter - stricter limits for authentication
	authLimiter, err := strigo.New(&strigo.Options{
		Points:      5,  // Only 5 login attempts
		Duration:    300, // per 5 minutes
		KeyPrefix:   "auth",
		StoreClient: createRedisClient(),
	})
	if err != nil {
		log.Printf("Failed to create auth limiter, using memory: %v", err)
		authLimiter, _ = strigo.New(&strigo.Options{
			Points:    5,
			Duration:  300,
			KeyPrefix: "auth",
		})
	}
	defer authLimiter.Close()

	// File upload limiter - expensive operations
	uploadLimiter, err := strigo.New(&strigo.Options{
		Points:      10, // 10 uploads
		Duration:    3600, // per hour
		KeyPrefix:   "upload",
		StoreClient: createRedisClient(),
	})
	if err != nil {
		log.Printf("Failed to create upload limiter, using memory: %v", err)
		uploadLimiter, _ = strigo.New(&strigo.Options{
			Points:    10,
			Duration:  3600,
			KeyPrefix: "upload",
		})
	}
	defer uploadLimiter.Close()

	// Middleware function to apply rate limiting
	rateLimitMiddleware := func(limiter *strigo.RateLimiter, points int64) fiber.Handler {
		return func(c *fiber.Ctx) error {
			// Use IP address as the key (you could also use user ID, API key, etc.)
			key := c.IP()
			
			result, err := limiter.Consume(key, points)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Rate limiter error",
				})
			}

			// Add rate limit headers (following HTTP standards)
			headers := result.Headers()
			for name, value := range headers {
				c.Set(name, value)
			}

			// Check if request is allowed
			if !result.Allowed {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error":       "Rate limit exceeded",
					"retryAfter":  result.MsBeforeNext / 1000,
					"remaining":   result.RemainingPoints,
					"limit":       result.TotalHits,
					"resetTime":   result.MsBeforeNext,
				})
			}

			return c.Next()
		}
	}

	// Public API endpoint - standard rate limiting
	app.Get("/api/data", rateLimitMiddleware(apiLimiter, 1), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Here's your data!",
			"data":    []string{"item1", "item2", "item3"},
		})
	})

	// Expensive API endpoint - consumes more points
	app.Get("/api/report", rateLimitMiddleware(apiLimiter, 5), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Here's your expensive report!",
			"report":  "This took a lot of resources to generate",
		})
	})

	// Authentication endpoint - strict rate limiting
	app.Post("/auth/login", rateLimitMiddleware(authLimiter, 1), func(c *fiber.Ctx) error {
		// Simulate login logic
		username := c.FormValue("username")
		password := c.FormValue("password")
		
		if username == "demo" && password == "password" {
			return c.JSON(fiber.Map{
				"success": true,
				"token":   "dummy-jwt-token",
			})
		}
		
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	})

	// File upload endpoint - resource-intensive operation
	app.Post("/upload", rateLimitMiddleware(uploadLimiter, 1), func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No file provided",
			})
		}

		return c.JSON(fiber.Map{
			"message":  "File uploaded successfully",
			"filename": file.Filename,
			"size":     file.Size,
		})
	})

	// Status endpoint to check rate limit status
	app.Get("/status/:endpoint", func(c *fiber.Ctx) error {
		endpoint := c.Params("endpoint")
		key := c.IP()

		var limiter *strigo.RateLimiter
		switch endpoint {
		case "api":
			limiter = apiLimiter
		case "auth":
			limiter = authLimiter
		case "upload":
			limiter = uploadLimiter
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Unknown endpoint",
			})
		}

		// Get current status without consuming points
		result, err := limiter.Get(key)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get status",
			})
		}

		if result == nil {
			return c.JSON(fiber.Map{
				"message": "No rate limit data",
				"endpoint": endpoint,
			})
		}

		return c.JSON(fiber.Map{
			"endpoint":         endpoint,
			"remaining":        result.RemainingPoints,
			"consumed":         result.ConsumedPoints,
			"limit":           result.TotalHits,
			"resetInMs":       result.MsBeforeNext,
			"isFirstInWindow": result.IsFirstInDuration,
		})
	})

	// Health check endpoint (no rate limiting)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"service": "strigo-demo",
		})
	})

	log.Println("🚀 Server starting on port 3000")
	log.Println("📊 Endpoints:")
	log.Println("  GET  /api/data         - Standard API (1 point)")
	log.Println("  GET  /api/report       - Expensive API (5 points)")
	log.Println("  POST /auth/login       - Authentication (strict limits)")
	log.Println("  POST /upload           - File upload (hourly limits)")
	log.Println("  GET  /status/:endpoint - Check rate limit status")
	log.Println("  GET  /health           - Health check (no limits)")
	
	log.Fatal(app.Listen(":3000"))
}

// createRedisClient creates a Redis client with fallback handling
func createRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}
