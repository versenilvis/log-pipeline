package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	db "github.com/versenilvis/log-pipeline/db/sqlc"
	"github.com/versenilvis/log-pipeline/internal/config"
	"github.com/versenilvis/log-pipeline/internal/logger"
	"github.com/versenilvis/log-pipeline/internal/tracing"
)

func main() {
	logger.InitLogger()
	defer func() {
		if err := logger.Log.Sync(); err != nil {
			log.Printf("Log sync error (may be benign): %v\n", err)
		}
	}()

	if err := godotenv.Load(); err != nil {
		logger.Log.Info("no .env file found, using deploy env vars")
	} else {
		logger.Log.Info("using .env file for configuration")
	}

	cfg := config.LoadConfig()
	tracing.InitTracingConfig(cfg.Tracing.HeaderTraceID, cfg.Tracing.HeaderSpanID)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		logger.Log.Fatal("failed to connect postgres", zap.Error(err))
	}
	defer pool.Close()

	queries := db.New(pool)

	hub := NewHub()
	go hub.Run()
	go StartListener(ctx, cfg, hub, queries)

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("query api ok")
	})
	app.Get("/v1/traces/:id", getTrace(queries))
	app.Get("/v1/logs", searchLogs(cfg, queries))
	app.Get("/v1/logs/stream", websocket.New(func(c *websocket.Conn) {
		client := &Client{hub: hub, conn: c, send: make(chan []byte, 16)}
		hub.register <- client
		go client.writePump()
		client.readPump()
	}))

	go func() {
		if err := app.Listen(":" + cfg.QueryPort); err != nil {
			logger.Log.Fatal("failed to start query server",
				zap.String("port", cfg.QueryPort), zap.Error(err))
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	logger.Log.Info("shutting down query api...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Log.Error("error during shutdown", zap.Error(err))
	}
}
