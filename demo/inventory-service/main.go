package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/versenilvis/log-pipeline/internal/config"
	"github.com/versenilvis/log-pipeline/internal/tracing"
)

func main() {
	_ = godotenv.Load()
	cfg := config.LoadConfig()

	reporter := tracing.NewReporter(cfg.DemoURLs.IngestURL, "inventory-service")

	app := fiber.New()
	app.Use(tracing.Middleware())

	app.Post("/reserve", func(c fiber.Ctx) error {
		start := time.Now()
		ctx := c.Context()

		reporter.Log(ctx, "info", "reserving inventory")

		time.Sleep(time.Duration(10+rand.Intn(40)) * time.Millisecond)

		status := "ok"
		if rand.Intn(30) == 0 { // ~3% errors
			status = "error"
			reporter.Log(ctx, "error", "insufficient stock")
		}

		reporter.Span(ctx, "reserve-inventory", status, int(time.Since(start).Milliseconds()))

		if status == "error" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "out of stock"})
		}
		return c.JSON(fiber.Map{"status": "reserved"})
	})

	if err := app.Listen(":9003"); err != nil {
		log.Fatalf("inventory-service server error: %v", err)
	}
}