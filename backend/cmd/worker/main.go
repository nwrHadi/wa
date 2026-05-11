package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wa-platform/backend/internal/config"
	"wa-platform/backend/internal/platform/cache"
	"wa-platform/backend/internal/platform/db"
	"wa-platform/backend/internal/platform/logger"
	"wa-platform/backend/internal/store"
	workerqueue "wa-platform/backend/internal/worker"
)

func main() {
	cfg := config.Load()
	log := logger.New("[worker]")
	ctx := context.Background()

	mariaDB, err := db.NewMariaDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect mariadb: %v", err)
	}
	defer mariaDB.Close()
	repo := store.NewRepository(mariaDB)

	redisClient, err := cache.NewRedis(ctx, cfg)
	if err != nil {
		if cfg.RedisOptional {
			log.Printf("redis unavailable, continuing in optional mode: %v", err)
			redisClient = nil
		} else {
			log.Fatalf("failed to connect redis: %v", err)
		}
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8090"
	}

	log.Printf("worker started with concurrency=%d", cfg.WorkerConcurrency)
	log.Printf("worker queue source=%s", fmt.Sprintf("redis://%s", cfg.RedisAddr))
	log.Printf("baileys gateway=%s", gatewayURL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	processor := workerqueue.Processor{Redis: redisClient, Log: log, Store: repo, Gateway: gatewayURL}
	go processor.Run(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("worker stopping")
	cancel()
	time.Sleep(300 * time.Millisecond)
}
