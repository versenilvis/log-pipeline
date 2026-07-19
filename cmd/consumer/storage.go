package main

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

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
