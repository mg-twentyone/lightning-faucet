package main

import (
	"log"
	"time"

	"faucet/internal/configs"
	"faucet/internal/handlers"
	"faucet/internal/security"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

func main() {
	configs.LoadConfig()

	// initialize security services
	claimTracker := security.NewRateLimiter()

	engine := html.New("./web/templates", ".html")
	engine.Delims("[[", "]]")

	// initialize Fiber app (express-inspired web framework)
	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
		},
	})

	app.Use(logger.New())

	// spam protection
	app.Use(limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{"detail": "Too many requests to the server."})
		},
	}))

	// routes
	app.Get("/", handlers.HandleIndex)
	app.Get("/api/info", handlers.HandleBalance)

	app.Post("/api/claim", func(c *fiber.Ctx) error {
		return handlers.HandleClaim(c, claimTracker)
	})

	// server start (for docker)
	log.Printf("⚡ Faucet starting on :8000 (Max Claim: %d sats)", configs.MaxClaimAmount)
	log.Fatal(app.Listen("0.0.0.0:8000"))
}