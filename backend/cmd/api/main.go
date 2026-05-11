package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wa-platform/backend/internal/config"
	"wa-platform/backend/internal/http/handlers"
	"wa-platform/backend/internal/http/server"
	"wa-platform/backend/internal/platform/cache"
	"wa-platform/backend/internal/platform/db"
	"wa-platform/backend/internal/platform/logger"
	"wa-platform/backend/internal/realtime"
	"wa-platform/backend/internal/service"
	"wa-platform/backend/internal/store"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	log := logger.New("[api]")

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
			redisClient = redis.NewClient(&redis.Options{
				Addr:     cfg.RedisAddr,
				Password: cfg.RedisPassword,
				DB:       cfg.RedisDB,
			})
		} else {
			log.Fatalf("failed to connect redis: %v", err)
		}
	}
	defer redisClient.Close()

	hub := realtime.NewHub()
	queue := service.NewQueue(redisClient)
	webhookService := service.NewWebhookService(repo, queue)
	eventService := service.NewEventService(repo, queue, hub, webhookService)
	scheduler := &service.Scheduler{Store: repo, Queue: queue, Webhook: webhookService, Log: log}
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8090"
	}
	go hub.Run(ctx)
	go scheduler.Run(ctx)

	h := server.Handlers{
		Health:         handlers.HealthHandler{DB: mariaDB, Redis: redisClient},
		Auth:           handlers.AuthHandler{Cfg: cfg, Store: repo},
		Dashboard:      handlers.DashboardHandler{Store: repo},
		Devices:        handlers.DevicesHandler{Store: repo, Queue: queue, Hub: hub},
		Messages:       handlers.MessagesHandler{Store: repo, Queue: queue, Hub: hub, GatewayURL: gatewayURL},
		Logs:           handlers.LogsHandler{Store: repo},
		Webhooks:       handlers.WebhookHandler{Store: repo, Queue: queue},
		InternalEvents: handlers.InternalEventsHandler{Events: eventService},
		Realtime:       handlers.RealtimeHandler{Hub: hub},
	}

	e := server.NewRouter(cfg, h)
	addr := fmt.Sprintf(":%s", cfg.APIPort)

	go func() {
		if err := e.Start(addr); err != nil {
			log.Printf("http server stopped: %v", err)
		}
	}()
	log.Printf("api listening on %s", addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}
