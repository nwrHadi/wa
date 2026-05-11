package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"wa-platform/backend/internal/store"
)

type Scheduler struct {
	Store   *store.Repository
	Queue   *Queue
	Webhook *WebhookService
	Log     *log.Logger
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.drainDueWebhookDeliveries(ctx)
		}
	}
}

func (s *Scheduler) drainDueWebhookDeliveries(ctx context.Context) {
	if s.Store == nil || s.Queue == nil || !s.Queue.Enabled() {
		return
	}
	deliveries, err := s.Store.DueWebhookDeliveries(ctx, 100)
	if err != nil {
		if s.Log != nil {
			s.Log.Printf("webhook schedule query error: %v", err)
		}
		return
	}
	for _, delivery := range deliveries {
		jobPayload, _ := json.Marshal(map[string]any{"webhookId": delivery.WebhookID, "eventId": delivery.EventID})
		_ = s.Queue.Enqueue(ctx, QueueJob{Type: WebhookDeliveryType, Payload: jobPayload, CreatedAt: time.Now().UTC()})
	}
}
