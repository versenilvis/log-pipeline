package main

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/versenilvis/log-pipeline/internal/logger"
	"github.com/versenilvis/log-pipeline/internal/models"
)

// move to config to use env var soon
const (
	maxBatchSize = 500
	streamName   = "ingest_stream"
	streamMaxLen = 100000
)

var (
	validLevels   = map[string]bool{"info": true, "warn": true, "error": true}
	validStatuses = map[string]bool{"ok": true, "error": true}
)

func handleIngest(rdb *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		var req models.IngestRequest

		// in my design, I only accept JSON request bodies
		// if you want to support XML, forms, messagepack, ..., change it to c.Bind().Body()
		// NOTE: https://docs.gofiber.io/next/api/bind
		// if err := c.Bind().Body(&req); err != nil {
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid JSON body",
			})
		}

		if len(req.Entries) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "entries must not be empty",
			})
		}
		if len(req.Entries) > maxBatchSize {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "batch too large, max " + strconv.Itoa(maxBatchSize),
			})
		}

		for i, e := range req.Entries {
			if err := validateEntry(e); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "entry " + strconv.Itoa(i) + ": " + err.Error(),
				})
			}
		}

		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		pipe := rdb.Pipeline()
		// gather all logs that need to be sent into the pipeline
		for _, e := range req.Entries {
			pipe.XAdd(ctx, &redis.XAddArgs{
				Stream: streamName,
				MaxLen: streamMaxLen,
				Approx: true,
				Values: entryToFields(e),
			})
		}

		// then send all logs in one time (1 RTT)
		if _, err := pipe.Exec(ctx); err != nil {
			logger.Log.Error("failed to push batch to redis stream", zap.Error(err))
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "failed to enqueue entries",
			})
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"accepted": len(req.Entries),
		})
	}
}
