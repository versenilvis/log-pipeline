package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Entry struct {
	Type         string          `json:"type"` // "log" | "span"
	TraceID      uuid.UUID       `json:"trace_id"`
	SpanID       *uuid.UUID      `json:"span_id,omitempty"`
	ParentSpanID *uuid.UUID      `json:"parent_span_id,omitempty"`
	Service      string          `json:"service"`
	Level        *string         `json:"level,omitempty"`
	Status       *string         `json:"status,omitempty"`
	SpanName     *string         `json:"span_name,omitempty"`
	Message      *string         `json:"message,omitempty"`
	DurationMs   *int            `json:"duration_ms,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	Timestamp    time.Time       `json:"timestamp"`
}

type IngestRequest struct {
	Entries []Entry `json:"entries"`
}
