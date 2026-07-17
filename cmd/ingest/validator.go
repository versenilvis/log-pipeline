// cmd/ingest/validate.go
package main

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/versenilvis/log-pipeline/internal/models"
)

func validateEntry(e models.Entry) error {
	if e.Service == "" {
		return errors.New("service is required")
	}
	if e.TraceID == uuid.Nil {
		return errors.New("trace_id is required")
	}
	if e.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}
	if e.Timestamp.After(time.Now().Add(5 * time.Minute)) {
		return errors.New("timestamp cannot be in the future")
	}

	switch e.Type {
	case "log":
		if e.Level == nil || !validLevels[*e.Level] {
			return errors.New(`level is required and must be "info|warn|error" when "type=log"`)
		}
	case "span":
		if e.DurationMs == nil {
			return errors.New(`duration_ms is required when "type=span"`)
		}
		if e.Status != nil && !validStatuses[*e.Status] {
			return errors.New(`status must be "ok|error" when set`)
		}
	default:
		return errors.New(`type must be "log|span"`)
	}

	return nil
}
