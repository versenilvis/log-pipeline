package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/versenilvis/log-pipeline/internal/models"
)

func requireString(values map[string]any, key string) (string, error) {
	v, ok := values[key].(string)
	if !ok || v == "" {
		return "", fmt.Errorf("missing %s", key)
	}
	return v, nil
}

func requireUUID(values map[string]any, key string) (uuid.UUID, error) {
	s, err := requireString(values, key)
	if err != nil {
		return uuid.UUID{}, err
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid %s: %w", key, err)
	}
	return id, nil
}

func optionalStringPtr(values map[string]any, key string) *string {
	if v, ok := values[key].(string); ok {
		return &v
	}
	return nil
}

func optionalUUIDPtr(values map[string]any, key string) (*uuid.UUID, error) {
	v, ok := values[key].(string)
	if !ok {
		return nil, nil
	}
	id, err := uuid.Parse(v)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", key, err)
	}
	return &id, nil
}

func optionalIntPtr(values map[string]any, key string) (*int, error) {
	v, ok := values[key].(string)
	if !ok {
		return nil, nil
	}
	d, err := strconv.Atoi(v)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", key, err)
	}
	return &d, nil
}

func parseFields(values map[string]any) (models.Entry, error) {
	var e models.Entry
	var err error

	if e.Type, err = requireString(values, "type"); err != nil {
		return e, err
	}
	if e.TraceID, err = requireUUID(values, "trace_id"); err != nil {
		return e, err
	}
	if e.Service, err = requireString(values, "service"); err != nil {
		return e, err
	}

	tsStr, err := requireString(values, "timestamp")
	if err != nil {
		return e, err
	}
	if e.Timestamp, err = time.Parse(time.RFC3339Nano, tsStr); err != nil {
		return e, fmt.Errorf("invalid timestamp: %w", err)
	}

	if e.SpanID, err = optionalUUIDPtr(values, "span_id"); err != nil {
		return e, err
	}
	if e.ParentSpanID, err = optionalUUIDPtr(values, "parent_span_id"); err != nil {
		return e, err
	}
	if e.DurationMs, err = optionalIntPtr(values, "duration_ms"); err != nil {
		return e, err
	}

	e.Level = optionalStringPtr(values, "level")
	e.Status = optionalStringPtr(values, "status")
	e.SpanName = optionalStringPtr(values, "span_name")
	e.Message = optionalStringPtr(values, "message")

	if v, ok := values["metadata"].(string); ok && v != "" {
		e.Metadata = json.RawMessage(v)
	}

	return e, nil
}
