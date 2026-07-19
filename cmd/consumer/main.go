package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/versenilvis/log-pipeline/internal/config"
	"github.com/versenilvis/log-pipeline/internal/logger"
)

func main() {
	logger.InitLogger()
	defer logger.Log.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Log.Info("no .env file found, using deploy env vars")
	} else {
		logger.Log.Info("using .env file for configuration")
	}

	cfg := config.LoadConfig()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Log.Fatal("failed to connect redis", zap.Error(err))
	}

	pool, err := pgxpool.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		logger.Log.Fatal("failed to connect postgres", zap.Error(err))
	}
	defer pool.Close()

	// The consumer name uses the pod's HOSTNAME (K8s automatically sets it),
	// not hardcoded, ensuring no duplicates when scaling multiple replicas,
	// and making it easy to trace which pod holds which entry when debugging
	consumerName, _ := os.Hostname()
	if consumerName == "" {
		consumerName = "consumer-unknown"
	}

	// MkStream ensures the consumer group exists, creating it if not
	// "$" means "start reading from the latest message" (not historical data)
	if err := rdb.XGroupCreateMkStream(ctx, streamName, groupName, "$").Err(); err != nil {
		if !strings.Contains(err.Error(), "BUSYGROUP") {
			logger.Log.Fatal("failed to create consumer group", zap.Error(err))
		}
	}

	c := NewConsumer(rdb, pool, consumerName)

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go c.RunMainLoop(appCtx)
	go c.RunRecoveryLoop(appCtx)

	sig := make(chan os.Signal, 1)                    // create a transmission channel to receive signals from the system
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM) // ctrl c or stop signal from k8s

	// sig forces main to wait, to let the two sub-threads do their job
	// this allows the two underlying sub-threads to run repeatedly and process logs continuously day after day until we proactively send a shutdown signal (such as pressing Ctrl+C)
	<-sig

	// after sig ended, print log, call cancel()
	logger.Log.Info("shutting down consumer...")
	cancel()
	// wait for 1 sec for the consumer to finish processing the last message
	time.Sleep(1 * time.Second)
}
