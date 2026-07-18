package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/versenilvis/log-pipeline/internal/config"
	"github.com/versenilvis/log-pipeline/internal/logger"
	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		logger.Log.Info("no .env file found, using deploy env vars")
	} else {
		logger.Log.Info("using .env file for configuration")
	}

	logger.InitLogger()
	defer func() {
		if err := logger.Log.Sync(); err != nil {
			log.Printf("Log sync error (may be benign): %v\n", err)
		}
	}()

	cfg := config.LoadConfig()
	logger.Log.Debug("Application Configuration Loaded", zap.String("port", cfg.Port))

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Log.Fatal("Failed to connect Redis", zap.Error(err))
	}

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("hello")
	})
	app.Post("/v1/ingest", handleIngest(redisClient))

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			logger.Log.Panic("Server failed to start", zap.Error(err))
		}
	}()

	// Ctrl-C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Log.Info("SHUTTING DOWN SERVER...")

	/* if there is a request in progress (e.g. waiting for a slow Redis response),
	the default Shutdown() function may wait indefinitely
	it's recommended to use a context with a timeout to ensure the shutdown occurs within a reasonable timeframe
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Log.Error("Error during shutdown", zap.Error(err))
	}
}
