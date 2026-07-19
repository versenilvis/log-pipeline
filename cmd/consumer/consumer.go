package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/versenilvis/log-pipeline/internal/logger"
)

// move to config to use env var soon
const (
	streamName     = "ingest_stream"
	groupName      = "log_processors"
	deadLetterName = "dead_letter_stream"
	batchSize      = 50
	blockTimeout   = 2 * time.Second
	idleThreshold  = 15 * time.Second
	recoveryTick   = 15 * time.Second
)

type Consumer struct {
	rdb  *redis.Client
	pool *pgxpool.Pool
	name string
}

func NewConsumer(rdb *redis.Client, pool *pgxpool.Pool, name string) *Consumer {
	return &Consumer{rdb: rdb, pool: pool, name: name}
}

func (c *Consumer) RunMainLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		// xreadgroup: command to wait for receiving messages
		// ">" means only messages that have not been assigned to anyone else will be read
		res, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: c.name,
			Streams:  []string{streamName, ">"},
			Count:    batchSize,
			Block:    blockTimeout,
			// we don't use block: 0 here to make sure the for loop go back to check ctx.Done() and also prevent long time blocking
		}).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			logger.Log.Error("XReadGroup failed", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}

		for _, stream := range res {
			c.processMessages(ctx, stream.Messages)
		}
	}
}

func (c *Consumer) RunRecoveryLoop(ctx context.Context) {
	ticker := time.NewTicker(recoveryTick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		// every recoveryTick asks recoverPending to check any message that stuck in the stream
		case <-ticker.C:
			c.recoverPending(ctx)
		}
	}
}

func (c *Consumer) recoverPending(ctx context.Context) {
	pending, err := c.rdb.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: streamName,
		Group:  groupName,
		Idle:   idleThreshold,
		Start:  "-",
		End:    "+",
		Count:  batchSize,
	}).Result()
	if err != nil {
		logger.Log.Error("XPendingExt failed", zap.Error(err))
		return
	}
	// if redis return len(pending) == 0 means no messages stuck in the stream
	if len(pending) == 0 {
		return
	}

	// if yes, ask redis to assign those messages to this consumer again
	ids := make([]string, len(pending))
	for i, p := range pending {
		ids[i] = p.ID
	}

	claimed, err := c.rdb.XClaim(ctx, &redis.XClaimArgs{
		Stream:   streamName,
		Group:    groupName,
		Consumer: c.name,
		MinIdle:  idleThreshold,
		Messages: ids,
	}).Result()
	if err != nil {
		logger.Log.Error("XClaim failed", zap.Error(err))
		return
	}

	logger.Log.Info("recovered pending entries", zap.Int("count", len(claimed)))
	c.processMessages(ctx, claimed)
}
