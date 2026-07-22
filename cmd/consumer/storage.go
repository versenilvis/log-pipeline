package main

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/versenilvis/log-pipeline/internal/logger"
	"github.com/versenilvis/log-pipeline/internal/models"
)

func bulkInsert(ctx context.Context, pool *pgxpool.Pool, entries []models.Entry) error {
	rows := make([][]any, len(entries))
	for i, e := range entries {
		rows[i] = []any{
			e.Type, e.TraceID, e.SpanID, e.ParentSpanID, e.Service,
			e.Level, e.Status, e.SpanName, e.Message, e.DurationMs,
			[]byte(e.Metadata), e.Timestamp,
		}
	}

	_, err := pool.CopyFrom(
		ctx,
		pgx.Identifier{"entries"},
		[]string{
			"type", "trace_id", "span_id", "parent_span_id", "service",
			"level", "status", "span_name", "message", "duration_ms",
			"metadata", "created_at",
		},
		pgx.CopyFromRows(rows),
	)
	return err
}

func notifyNewEntries(ctx context.Context, pool *pgxpool.Pool, entries []models.Entry, payloadLimit int) {
	payload, err := json.Marshal(entries)
	if err != nil {
		logger.Log.Warn("failed to marshal entries for notify", zap.Error(err))
		return
	}

	payloadStr := string(payload)
	if payloadLimit > 0 && len(payloadStr) > payloadLimit {
		payloadStr = ""
	}

	_, err = pool.Exec(ctx, "SELECT pg_notify('new_entry', $1)", payloadStr)
	if err != nil {
		logger.Log.Warn("failed to notify new entry", zap.Error(err))
	}
}
