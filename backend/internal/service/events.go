package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"wa-platform/backend/internal/realtime"
	"wa-platform/backend/internal/store"
)

type EventService struct {
	Store   *store.Repository
	Queue   *Queue
	Hub     *realtime.Hub
	Webhook *WebhookService
}

type EventEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type DeviceEventPayload struct {
	DeviceKey   string          `json:"deviceKey"`
	Label       string          `json:"label,omitempty"`
	Status      string          `json:"status"`
	PhoneNumber string          `json:"phoneNumber,omitempty"`
	QRText      string          `json:"qrText,omitempty"`
	SessionBlob json.RawMessage `json:"sessionBlob,omitempty"`
	LastSeenAt  *time.Time      `json:"lastSeenAt,omitempty"`
}

type MessageEventPayload struct {
	DeviceKey     string `json:"deviceKey"`
	ExternalRefID string `json:"externalRefId"`
	WAMessageID   string `json:"waMessageId,omitempty"`
	ToNumber      string `json:"toNumber,omitempty"`
	Body          string `json:"body,omitempty"`
	Status        string `json:"status"`
	ErrorMessage  string `json:"errorMessage,omitempty"`
}

type WebhookDeliveryPayload struct {
	DeliveryID uint64 `json:"deliveryId"`
}

func NewEventService(store *store.Repository, queue *Queue, hub *realtime.Hub, webhook *WebhookService) *EventService {
	return &EventService{Store: store, Queue: queue, Hub: hub, Webhook: webhook}
}

func (s *EventService) Handle(ctx context.Context, envelope EventEnvelope) error {
	switch envelope.Type {
	case "device.connected", "device.disconnected", "device.qr_updated", "device.qr.updated":
		return s.handleDeviceEvent(ctx, envelope)
	case "message.processing", "message.sent", "message.failed", "message.received":
		return s.handleMessageEvent(ctx, envelope)
	case "webhook.delivery.succeeded", "webhook.delivery.failed":
		return s.handleWebhookDeliveryEvent(ctx, envelope)
	default:
		s.publish("event.unknown", map[string]any{"type": envelope.Type})
		return nil
	}
}

func (s *EventService) handleDeviceEvent(ctx context.Context, envelope EventEnvelope) error {
	var payload DeviceEventPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return err
	}
	if payload.Status == "" {
		switch envelope.Type {
		case "device.connected":
			payload.Status = "connected"
		case "device.disconnected":
			payload.Status = "disconnected"
		case "device.connecting", "device.reconnecting", "device.qr_updated", "device.qr.updated":
			payload.Status = "connecting"
		default:
			payload.Status = "disconnected"
		}
	}
	if payload.Label == "" {
		payload.Label = payload.DeviceKey
	}
	device, err := s.Store.UpsertDevice(ctx, payload.DeviceKey, payload.Label, payload.Status, payload.PhoneNumber, payload.QRText, payload.SessionBlob, payload.LastSeenAt)
	if err != nil {
		return err
	}
	s.publish(envelope.Type, device)
	return nil
}

func (s *EventService) handleMessageEvent(ctx context.Context, envelope EventEnvelope) error {
	var payload MessageEventPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return err
	}
	message, err := s.Store.MessageByExternalRef(ctx, payload.ExternalRefID)
	if err != nil {
		return err
	}
	if payload.Status != "" {
		if err := s.Store.UpdateMessageStatus(ctx, payload.ExternalRefID, payload.Status, payload.WAMessageID, payload.ErrorMessage); err != nil {
			return err
		}
		message, err = s.Store.MessageByExternalRef(ctx, payload.ExternalRefID)
		if err != nil {
			return err
		}
	}
	if err := s.Store.InsertMessageEvent(ctx, message.ID, envelope.Type, payload); err != nil {
		return err
	}
	s.publish(envelope.Type, message)
	if s.Webhook != nil {
		_ = s.Webhook.Dispatch(ctx, envelope.Type, message)
	}
	return nil
}

func (s *EventService) handleWebhookDeliveryEvent(ctx context.Context, envelope EventEnvelope) error {
	var payload WebhookDeliveryPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return err
	}
	s.publish(envelope.Type, payload)
	return nil
}

func (s *EventService) publish(eventType string, payload any) {
	if s.Hub == nil {
		return
	}
	s.Hub.Publish(realtime.Event{Type: eventType, Payload: payload})
}

func NewExternalRefID() string {
	buf := make([]byte, 12)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func BuildEventID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
