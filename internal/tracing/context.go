package tracing

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey string

// TODO: config
const (
	traceIDKey      ctxKey = "trace_id"
	spanIDKey       ctxKey = "span_id"
	parentSpanIDKey ctxKey = "parent_span_id"

	HeaderTraceID = "X-Trace-Id"
	HeaderSpanID  = "X-Span-Id"
)

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
