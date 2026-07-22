// Package config provides the configuration for the application
package config

import (
	"fmt"
	"time"

	"github.com/versenilvis/log-pipeline/internal/utils"
)

type AppConfig struct {
	QueryPort       string          `json:"query_port"`
	IngestPort      string          `json:"ingest_port"`
	Consumer        ConsumerConfig  `json:"consumer_config"`
	Tracing         TracingConfig   `json:"tracing_config"`
	Query           QueryConfig     `json:"query_config"`
	DemoURLs        DemoServiceURLs `json:"demo_urls"`
	Redis           RedisConfig     `json:"redis_config"`
	Postgres        PostgresConfig  `json:"postgres_config"`
	LogSamplingRate int             `json:"log_sampling_rate"`
}

type ConsumerConfig struct {
	StreamName         string        `json:"stream_name"`
	GroupName          string        `json:"group_name"`
	DeadLetterName     string        `json:"dead_letter_name"`
	BatchSize          int64         `json:"batch_size"`
	BlockTimeout       time.Duration `json:"block_timeout"`
	IdleThreshold      time.Duration `json:"idle_threshold"`
	RecoveryTick       time.Duration `json:"recovery_tick"`
	NotifyPayloadLimit int           `json:"notify_payload_limit"`
}

type TracingConfig struct {
	HeaderTraceID string `json:"header_trace_id"`
	HeaderSpanID  string `json:"header_span_id"`
}

type QueryConfig struct {
	DefaultPageSize       int32 `json:"default_page_size"`
	MaxPageSize           int32 `json:"max_page_size"`
	ListenerFallbackLimit int32 `json:"listener_fallback_limit"`
}

type DemoServiceURLs struct {
	IngestURL           string `json:"ingest_url"`
	OrderServiceURL     string `json:"order_service_url"`
	PaymentServiceURL   string `json:"payment_service_url"`
	InventoryServiceURL string `json:"inventory_service_url"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"-"` // sensitive information please use "-"
	DB       int    `json:"db"`
}

type PostgresConfig struct {
	DSN string `json:"-"`
}

// type Duration time.Duration

// func (d Duration) MarshalJSON() ([]byte, error) {
// 	return []byte(`"` + time.Duration(d).String() + `"`), nil
// }

func LoadConfig() *AppConfig {
	dbUser := utils.GetEnv("DB_USER", "postgres")
	dbPass := utils.GetEnv("DB_PASSWORD", "postgres")
	dbName := utils.GetEnv("DB_NAME", "logpipeline")
	dbPort := utils.GetEnv("DB_PORT", "5432")
	dbHost := utils.GetEnv("DB_HOST", "localhost")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	cfg := &AppConfig{
		QueryPort:  utils.GetEnv("QUERY_PORT", "8081"),
		IngestPort: utils.GetEnv("INGEST_PORT", "8080"),
		Consumer: ConsumerConfig{
			StreamName:         utils.GetEnv("REDIS_STREAM_NAME", "ingest_stream"),
			GroupName:          utils.GetEnv("REDIS_GROUP_NAME", "log_processors"),
			DeadLetterName:     utils.GetEnv("REDIS_DEAD_LETTER_NAME", "dead_letter_stream"),
			BatchSize:          int64(utils.GetEnvAsInt("CONSUMER_BATCH_SIZE", 50)),
			BlockTimeout:       utils.GetEnvAsDuration("CONSUMER_BLOCK_TIMEOUT", 2*time.Second),
			IdleThreshold:      utils.GetEnvAsDuration("CONSUMER_IDLE_THRESHOLD", 15*time.Second),
			RecoveryTick:       utils.GetEnvAsDuration("CONSUMER_RECOVERY_TICK", 15*time.Second),
			NotifyPayloadLimit: utils.GetEnvAsInt("NOTIFY_PAYLOAD_LIMIT", 7500),
		},
		Tracing: TracingConfig{
			HeaderTraceID: utils.GetEnv("HEADER_TRACE_ID", "X-Trace-Id"),
			HeaderSpanID:  utils.GetEnv("HEADER_SPAN_ID", "X-Span-Id"),
		},
		Query: QueryConfig{
			DefaultPageSize:       int32(utils.GetEnvAsInt("QUERY_DEFAULT_PAGE_SIZE", 50)),
			MaxPageSize:           int32(utils.GetEnvAsInt("QUERY_MAX_PAGE_SIZE", 200)),
			ListenerFallbackLimit: int32(utils.GetEnvAsInt("QUERY_LISTENER_FALLBACK_LIMIT", 20)),
		},
		DemoURLs: DemoServiceURLs{
			IngestURL:           utils.GetEnv("INGEST_URL", "http://localhost:8080"),
			OrderServiceURL:     utils.GetEnv("ORDER_SERVICE_URL", "http://localhost:9001"),
			PaymentServiceURL:   utils.GetEnv("PAYMENT_SERVICE_URL", "http://localhost:9002"),
			InventoryServiceURL: utils.GetEnv("INVENTORY_SERVICE_URL", "http://localhost:9003"),
		},
		Redis: RedisConfig{
			Addr:     utils.GetEnv("REDIS_ADDR", ""),
			Password: utils.GetEnv("REDIS_PASS", ""),
			DB:       utils.GetEnvAsInt("REDIS_DB", 0),
		},
		Postgres: PostgresConfig{
			DSN: dsn,
		},
		LogSamplingRate: utils.GetEnvAsInt("LOG_SAMPLING_RATE", 10),
	}
	return cfg
}
