package tracing

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey string

const (
	traceIDKey      ctxKey = "trace_id"
	spanIDKey       ctxKey = "span_id"
	parentSpanIDKey ctxKey = "parent_span_id"
)

var (
	HeaderTraceID = "X-Trace-Id"
	HeaderSpanID  = "X-Span-Id"
)

func InitTracingConfig(headerTraceID, headerSpanID string) {
	if headerTraceID != "" {
		HeaderTraceID = headerTraceID
	}
	if headerSpanID != "" {
		HeaderSpanID = headerSpanID
	}
}

func TraceIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(traceIDKey).(uuid.UUID)
	return id
}

func SpanIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(spanIDKey).(uuid.UUID)
	return id
}

func ParentSpanIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(parentSpanIDKey).(string)
	return id
}
