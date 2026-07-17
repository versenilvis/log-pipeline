package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/versenilvis/log-pipeline/internal/models"
)

func TestValidateEntry(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		entry   models.Entry
		wantErr bool
	}{
		{
			name: "valid log entry",
			entry: models.Entry{
				Type: "log", Service: "svc", TraceID: uuid.New(),
				Level: new("info"), Timestamp: now,
			},
			wantErr: false,
		},
		{
			name: "valid span entry",
			entry: models.Entry{
				Type: "span", Service: "svc", TraceID: uuid.New(),
				DurationMs: new(120), Status: new("ok"), Timestamp: now,
			},
			wantErr: false,
		},
		{
			name:    "empty service",
			entry:   models.Entry{Type: "log", Service: "", TraceID: uuid.New(), Level: new("info"), Timestamp: now},
			wantErr: true,
		},
		{
			name:    "empty trace_id",
			entry:   models.Entry{Type: "log", Service: "svc", TraceID: uuid.Nil, Level: new("info"), Timestamp: now},
			wantErr: true,
		},
		{
			name:    "zero timestamp",
			entry:   models.Entry{Type: "log", Service: "svc", TraceID: uuid.New(), Level: new("info"), Timestamp: time.Time{}},
			wantErr: true,
		},
		{
			name:    "timestamp too far in future",
			entry:   models.Entry{Type: "log", Service: "svc", TraceID: uuid.New(), Level: new("info"), Timestamp: now.Add(1 * time.Hour)},
			wantErr: true,
		},
		{
			name:    "log missing level",
			entry:   models.Entry{Type: "log", Service: "svc", TraceID: uuid.New(), Timestamp: now},
			wantErr: true,
		},
		{
			name:    "log invalid level enum",
			entry:   models.Entry{Type: "log", Service: "svc", TraceID: uuid.New(), Level: new("banana"), Timestamp: now},
			wantErr: true,
		},
		{
			name:    "span missing duration_ms",
			entry:   models.Entry{Type: "span", Service: "svc", TraceID: uuid.New(), Timestamp: now},
			wantErr: true,
		},
		{
			name: "span invalid status enum",
			entry: models.Entry{
				Type: "span", Service: "svc", TraceID: uuid.New(),
				DurationMs: new(50), Status: new("banana"), Timestamp: now,
			},
			wantErr: true,
		},
		{
			name: "span without status is still valid (status optional)",
			entry: models.Entry{
				Type: "span", Service: "svc", TraceID: uuid.New(),
				DurationMs: new(50), Timestamp: now,
			},
			wantErr: false,
		},
		{
			name:    "invalid type",
			entry:   models.Entry{Type: "banana", Service: "svc", TraceID: uuid.New(), Timestamp: now},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEntry(tt.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
