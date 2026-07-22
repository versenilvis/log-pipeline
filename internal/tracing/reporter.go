package tracing

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type entryPayload struct {
	Type         string    `json:"type"`
	TraceID      string    `json:"trace_id"`
	SpanID       string    `json:"span_id,omitempty"`
	ParentSpanID string    `json:"parent_span_id,omitempty"`
	Service      string    `json:"service"`
	Level        string    `json:"level,omitempty"`
	Status       string    `json:"status,omitempty"`
	SpanName     string    `json:"span_name,omitempty"`
	Message      string    `json:"message,omitempty"`
	DurationMs   int       `json:"duration_ms,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

type Reporter struct {
	ingestURL string
	service   string
}

func NewReporter(ingestURL, service string) *Reporter {
	return &Reporter{ingestURL: ingestURL, service: service}
}

// NOTE: send() is "fire-and-forget" (does not handle network errors)
// acceptable for a demo service since its purpose is only to generate test traces,
// not requiring the high reliability of the pipeline itself
func (r *Reporter) send(ctx context.Context, entries []entryPayload) {
	body, err := json.Marshal(map[string]any{"entries": entries})
	if err != nil {
		return
	}
	req, err := http.NewRequestWithContext(ctx, "POST", r.ingestURL+"/v1/ingest", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req) // fire-and-forget (send and then continue, no need to wait for a response)
	if err == nil {
		_ = resp.Body.Close()
	}
}

func (r *Reporter) Log(ctx context.Context, level, message string) {
	r.send(ctx, []entryPayload{{
		Type: "log", TraceID: TraceIDFromContext(ctx).String(),
		Service: r.service, Level: level, Message: message,
		Timestamp: time.Now(),
	}})
}

func (r *Reporter) Span(ctx context.Context, spanName, status string, durationMs int) {
	r.send(ctx, []entryPayload{{
		Type: "span", TraceID: TraceIDFromContext(ctx).String(),
		SpanID:       SpanIDFromContext(ctx).String(),
		ParentSpanID: ParentSpanIDFromContext(ctx),
		Service:      r.service, SpanName: spanName, Status: status,
		DurationMs: durationMs, Timestamp: time.Now(),
	}})
}
