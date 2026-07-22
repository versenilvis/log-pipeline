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

	reporter := tracing.NewReporter(cfg.DemoURLs.IngestURL, "payment-service")

	app := fiber.New()
	app.Use(tracing.Middleware())

	app.Post("/charge", func(c fiber.Ctx) error {
		start := time.Now()
		ctx := c.Context()

		reporter.Log(ctx, "info", "processing payment charge")

		// simulate payment processing delay
		time.Sleep(time.Duration(20+rand.Intn(80)) * time.Millisecond)

		// simulate occasional errors to create status=error data on the waterfall
		status := "ok"
		if rand.Intn(20) == 0 { // ~5% errors
			status = "error"
			reporter.Log(ctx, "error", "payment gateway timeout")
		}

		reporter.Span(ctx, "process-payment", status, int(time.Since(start).Milliseconds()))

		if status == "error" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "payment failed"})
		}
		return c.JSON(fiber.Map{"status": "charged"})
	})

	if err := app.Listen(":9002"); err != nil {
		log.Fatalf("payment-service server error: %v", err)
	}
}