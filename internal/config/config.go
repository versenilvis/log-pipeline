// Package config provides the configuration for the application
package config

import (
	"fmt"

	"github.com/versenilvis/log-pipeline/internal/utils"
)

type AppConfig struct {
	QueryPort       string         `json:"query_port"`
	IngestPort      string         `json:"ingest_port"`
	Redis           RedisConfig    `json:"redis_config"`
	Postgres        PostgresConfig `json:"postgres_config"`
	LogSamplingRate int            `json:"log_sampling_rate"`
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
