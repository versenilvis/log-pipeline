CREATE TABLE entries (
  id BIGSERIAL,
  type VARCHAR(4) NOT NULL,          -- 'log' | 'span'
  trace_id UUID NOT NULL,
  span_id UUID,
  parent_span_id UUID,

  service VARCHAR(64) NOT NULL,

  -- Status and message
  level VARCHAR(10),                 -- 'info'|'warn'|'error', only type=log
  status VARCHAR(10),                -- 'ok' | 'error', only type=span
  span_name VARCHAR(128),            -- e.g: process-payment, query-db, only type=span
  message TEXT,                      -- only type=log

  duration_ms INT,                   -- only type=span
  metadata JSONB,                    -- for extra column in the future (service_instance/pod_name, user_id, order_id...)

  created_at TIMESTAMPTZ NOT NULL,

  PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Default partition: receives data if the worker fails to create a new partition on a given day.
CREATE TABLE entries_default PARTITION OF entries DEFAULT;

-- e.g.
-- CREATE TABLE entries_20260715 PARTITION OF entries
--     FOR VALUES FROM ('2026-07-15 00:00:00+07') TO ('2026-07-16 00:00:00+07');

CREATE INDEX idx_trace_id ON entries(trace_id);
CREATE INDEX idx_service_created ON entries(service, created_at DESC);
CREATE INDEX idx_metadata_gin ON entries USING GIN(metadata);
