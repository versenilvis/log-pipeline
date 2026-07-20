package main

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/versenilvis/log-pipeline/db/sqlc"
)

func getTrace(q *db.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		// take the text string located at the :id position in the URL,
		// check it, and convert it to a true UUID
		// if the string is not in the correct UUID format (e.g., misspelled, missing characters),
		// return an error immediately
		traceID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid trace_id"})
		}

		pgTraceID := pgtype.UUID{Bytes: traceID, Valid: true}
		entries, err := q.GetEntriesByTraceID(c.Context(), pgTraceID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "query failed"})
		}

		return c.JSON(fiber.Map{"entries": entries})
	}
}

func searchLogs(q *db.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		params := db.SearchLogsParams{
			// TODO: config
			Limit: 50, // default page size
		}

		if v := c.Query("service"); v != "" {
			params.Service = pgtype.Text{String: v, Valid: true}
		}
		if v := c.Query("level"); v != "" {
			params.Level = pgtype.Text{String: v, Valid: true}
		}
		if v := c.Query("q"); v != "" {
			params.Query = pgtype.Text{String: v, Valid: true}
		}
		if v := c.Query("from"); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid from timestamp"})
			}
			params.From = pgtype.Timestamptz{Time: t, Valid: true}
		}
		if v := c.Query("to"); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid to timestamp"})
			}
			params.To = pgtype.Timestamptz{Time: t, Valid: true}
		}
		if v := c.Query("before_id"); v != "" {
			id, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid before_id"})
			}
			params.BeforeID = pgtype.Int8{Int64: id, Valid: true}
		}
		if v := c.Query("limit"); v != "" {
			limit, err := strconv.Atoi(v)
			if err != nil || limit <= 0 || limit > 200 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid limit, must be between 1 and 200"})
			}
			params.Limit = int32(limit)
		}

		logs, err := q.SearchLogs(c.Context(), params)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "query failed"})
		}

		return c.JSON(fiber.Map{"logs": logs})
	}
}
