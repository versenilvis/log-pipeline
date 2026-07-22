package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/versenilvis/log-pipeline/internal/tracing"
)

func main() {
	ingestURL := "http://localhost:8080"
	reporter := tracing.NewReporter(ingestURL, "api-gateway")

	app := fiber.New()
	app.Use(tracing.Middleware())

	app.Post("/checkout", func(c fiber.Ctx) error {
		start := time.Now()
		ctx := c.Context()

		reporter.Log(ctx, "info", "received checkout request")

		req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:9001/orders", nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create request"})
		}
		tracing.InjectHeaders(ctx, req)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
		}

		status := "ok"
		if err != nil || (resp != nil && resp.StatusCode >= 400) {
			status = "error"
		}
		reporter.Span(ctx, "handle-checkout", status, int(time.Since(start).Milliseconds()))

		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "order-service unreachable"})
		}
		return c.JSON(fiber.Map{"status": "checkout received"})
	})

	if err := app.Listen(":9000"); err != nil {
		log.Fatalf("api-gateway server error: %v", err)
	}
}
