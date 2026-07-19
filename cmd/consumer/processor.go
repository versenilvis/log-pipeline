package main

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/versenilvis/log-pipeline/internal/logger"
	"github.com/versenilvis/log-pipeline/internal/models"
)

func (c *Consumer) processMessages(ctx context.Context, msgs []redis.XMessage) {
	if len(msgs) == 0 {
		return
	}

	// test crash
	// logger.Log.Info("CRASH TEST: Sleeping 10s. Press Ctrl+C NOW to kill consumer!")
	// time.Sleep(10 * time.Second)

	var validEntries []models.Entry
	var validIDs []string

	for _, msg := range msgs {
		entry, err := parseFields(msg.Values)
		if err != nil {
			logger.Log.Warn("failed to parse entry, sending to dead-letter",
				zap.String("id", msg.ID), zap.Error(err))
			c.sendToDeadLetter(ctx, msg)
			validIDs = append(validIDs, msg.ID)
			continue
		}
		validEntries = append(validEntries, entry)
		validIDs = append(validIDs, msg.ID)
	}

	if len(validEntries) > 0 {
		if err := bulkInsert(ctx, c.pool, validEntries); err != nil {
			logger.Log.Error("bulk insert failed, will retry via recovery", zap.Error(err))
			return
		}
	}

	if err := c.rdb.XAck(ctx, streamName, groupName, validIDs...).Err(); err != nil {
		logger.Log.Error("XAck failed", zap.Error(err))
	}
}

func (c *Consumer) sendToDeadLetter(ctx context.Context, msg redis.XMessage) {
	values := msg.Values
	values["_original_id"] = msg.ID
	c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: deadLetterName,
		Values: values,
	})
}
