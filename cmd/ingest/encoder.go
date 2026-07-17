package main

import (
	"strconv"
	"time"

	"github.com/versenilvis/log-pipeline/internal/models"
)

func entryToFields(e models.Entry) map[string]any {
	fields := map[string]any{
		"type":      e.Type,
		"trace_id":  e.TraceID.String(),
		"service":   e.Service,
		"timestamp": e.Timestamp.Format(time.RFC3339Nano),
	}

	if e.SpanID != nil {
		fields["span_id"] = e.SpanID.String()
	}
	if e.ParentSpanID != nil {
		fields["parent_span_id"] = e.ParentSpanID.String()
	}
	if e.Level != nil {
		fields["level"] = *e.Level
	}
	if e.Status != nil {
		fields["status"] = *e.Status
	}
	if e.SpanName != nil {
		fields["span_name"] = *e.SpanName
	}
	if e.Message != nil {
		fields["message"] = *e.Message
	}
	if e.DurationMs != nil {
		fields["duration_ms"] = strconv.Itoa(*e.DurationMs)
	}
	if len(e.Metadata) > 0 {
		fields["metadata"] = string(e.Metadata)
	}

	return fields
}
