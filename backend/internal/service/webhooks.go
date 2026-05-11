package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"wa-platform/backend/internal/store"
)

type WebhookService struct {
	Store      *store.Repository
	Queue      *Queue
	HTTPClient *http.Client
}

type webhookEventPayload struct {
	EventType string      `json:"eventType"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

func NewWebhookService(store *store.Repository, queue *Queue) *WebhookService {
	return &WebhookService{
		Store:      store,
		Queue:      queue,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *WebhookService) Dispatch(ctx context.Context, eventType string, data any) error {
	webhooks, err := s.Store.ListWebhooks(ctx)
	if err != nil {
		return err
	}

	payload := webhookEventPayload{
		EventType: eventType,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	for _, webhook := range webhooks {
		if !webhook.Enabled {
			continue
		}
		if !matchesEventFilter(webhook.EventFilters, eventType) {
			continue
		}
		eventID := BuildEventID("wh")
		now := time.Now().UTC()
		_, err := s.Store.InsertWebhookDelivery(ctx, webhook.ID, eventID, eventType, string(payloadJSON), "pending", 0, &now, "")
		if err != nil {
			continue
		}
		if s.Queue != nil && s.Queue.Enabled() {
			jobPayload := map[string]any{"webhookId": webhook.ID, "eventId": eventID}
			jobBytes, _ := json.Marshal(jobPayload)
			_ = s.Queue.Enqueue(ctx, QueueJob{Type: WebhookDeliveryType, Payload: jobBytes, CreatedAt: time.Now().UTC()})
		}
	}
	return nil
}

func (s *WebhookService) Deliver(ctx context.Context, delivery store.WebhookDelivery) error {
	webhooks, err := s.Store.ListWebhooks(ctx)
	if err != nil {
		return err
	}
	var webhook *store.Webhook
	for i := range webhooks {
		if webhooks[i].ID == delivery.WebhookID {
			webhook = &webhooks[i]
			break
		}
	}
	if webhook == nil {
		return fmt.Errorf("webhook not found")
	}

	body := []byte(delivery.Payload)
	signature := signWebhookPayload(webhook.Secret, body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.TargetURL, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", delivery.EventType)
	req.Header.Set("X-Webhook-ID", delivery.EventID)

	resp, err := s.HTTPClient.Do(req)
	if err != nil || resp.StatusCode >= 300 {
		lastErr := "delivery failed"
		if err != nil {
			lastErr = err.Error()
		}
		return fmt.Errorf(lastErr)
	}
	if resp.Body != nil {
		_ = resp.Body.Close()
	}
	return nil
}

func signWebhookPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func matchesEventFilter(filterJSON string, eventType string) bool {
	if strings.TrimSpace(filterJSON) == "" {
		return true
	}
	if strings.Contains(filterJSON, eventType) {
		return true
	}
	return false
}
