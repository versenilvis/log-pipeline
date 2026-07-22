package main

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	db "github.com/versenilvis/log-pipeline/db/sqlc"
	"github.com/versenilvis/log-pipeline/internal/logger"
)

func StartListener(ctx context.Context, dsn string, hub *Hub, q *db.Queries) {
	/*
	Open a single Postgres connection completely outside the connection pool
	(because the pool shares a connection for multiple queries, so the LISTEN command cannot be kept fixed)
	*/
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		logger.Log.Fatal("listener: failed to connect", zap.Error(err))
	}
	defer func() {
		_ = conn.Close(ctx)
	}()

	if _, err := conn.Exec(ctx, "LISTEN new_entry"); err != nil {
		logger.Log.Fatal("listener: failed to LISTEN", zap.Error(err))
	}

	logger.Log.Info("listening for new_entry notifications")

	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			logger.Log.Error("listener: wait failed", zap.Error(err))
			return
		}

		// payload contains the exact new entries in this batch!
		if notification.Payload != "" {
			hub.broadcast <- []byte(notification.Payload)
			continue
		}

		// payload was empty (e.g. payload size exceeded limit), query latest 20
		logs, err := q.SearchLogs(ctx, db.SearchLogsParams{Limit: 20})
		if err != nil {
			logger.Log.Error("listener: query failed", zap.Error(err))
			continue
		}
		// compresses them into JSON
		payload, err := json.Marshal(logs)
		if err != nil {
			logger.Log.Error("listener: marshal failed", zap.Error(err))
			continue
		}
		// and sends them directly to the hub.broadcast channel
		hub.broadcast <- payload
	}
}
