package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/versenilvis/log-pipeline/internal/config"
	"github.com/versenilvis/log-pipeline/internal/tracing"
)

func main() {
	_ = godotenv.Load()
	cfg := config.LoadConfig()

	reporter := tracing.NewReporter(cfg.DemoURLs.IngestURL, "order-service")

	app := fiber.New()
	app.Use(tracing.Middleware())

	app.Post("/orders", func(c fiber.Ctx) error {
		start := time.Now()
		ctx := c.Context()

		reporter.Log(ctx, "info", "processing order")

		payReq, err := http.NewRequestWithContext(ctx, "POST", cfg.DemoURLs.PaymentServiceURL+"/charge", nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create pay request"})
		}
		tracing.InjectHeaders(ctx, payReq)
		payResp, payErr := http.DefaultClient.Do(payReq)
		if payErr == nil {
			defer payResp.Body.Close()
		}

		invReq, err := http.NewRequestWithContext(ctx, "POST", cfg.DemoURLs.InventoryServiceURL+"/reserve", nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create inv request"})
		}
		tracing.InjectHeaders(ctx, invReq)
		invResp, invErr := http.DefaultClient.Do(invReq)
		if invErr == nil {
			defer invResp.Body.Close()
		}

		status := "ok"
		if payErr != nil || invErr != nil ||
			(payResp != nil && payResp.StatusCode >= 400) ||
			(invResp != nil && invResp.StatusCode >= 400) {
			status = "error"
		}
		reporter.Span(ctx, "process-order", status, int(time.Since(start).Milliseconds()))

		return c.JSON(fiber.Map{"status": "order processed"})
	})

	if err := app.Listen(":9001"); err != nil {
		log.Fatalf("order-service server error: %v", err)
	}
}
