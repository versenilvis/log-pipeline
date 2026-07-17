// Package config provides the configuration for the application
package config

import (
	"github.com/versenilvis/log-pipeline/internal/utils"
)

type AppConfig struct {
	Port            string      `json:"port"`
	Redis           RedisConfig `json:"redis_config"`
	LogSamplingRate int         `json:"log_sampling_rate"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"-"` // sensitive information please use "-"
	DB       int    `json:"db"`
}

// type Duration time.Duration

// func (d Duration) MarshalJSON() ([]byte, error) {
// 	return []byte(`"` + time.Duration(d).String() + `"`), nil
// }

func LoadConfig() *AppConfig {
	cfg := &AppConfig{
		Port: utils.GetEnv("PORT", "8080"),
		Redis: RedisConfig{
			Addr:     utils.GetEnv("REDIS_ADDR", ""),
			Password: utils.GetEnv("REDIS_PASS", ""),
			DB:       utils.GetEnvAsInt("REDIS_DB", 0),
		},
		LogSamplingRate: utils.GetEnvAsInt("LOG_SAMPLING_RATE", 10),
	}

	return cfg
}
