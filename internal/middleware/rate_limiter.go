package middleware

import (
	"log/slog"

	"temp_backend/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// NewGlobalRateLimiter returns a middleware that limits all requests per IP.
func NewGlobalRateLimiter(cfg config.Config, log *slog.Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.RateLimit.GlobalMax,
		Expiration: cfg.RateLimit.GlobalWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			log.Warn("global rate limit hit",
				slog.String("ip", c.IP()),
				slog.String("path", c.Path()),
				slog.String("method", c.Method()),
			)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":  "Too many requests. Please try again later.",
				"code":   "RATE_LIMITED",
				"status": fiber.StatusTooManyRequests,
			})
		},
	})
}

// NewAuthRateLimiter returns a stricter middleware for unauthenticated auth endpoints.
func NewAuthRateLimiter(cfg config.Config, log *slog.Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.RateLimit.AuthMax,
		Expiration: cfg.RateLimit.AuthWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			log.Warn("auth rate limit hit",
				slog.String("ip", c.IP()),
				slog.String("path", c.Path()),
				slog.String("method", c.Method()),
			)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":  "Too many attempts. Please try again later.",
				"code":   "RATE_LIMITED",
				"status": fiber.StatusTooManyRequests,
			})
		},
	})
}
