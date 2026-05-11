package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"wa-platform/backend/internal/store"

	"github.com/redis/go-redis/v9"
)

type Job struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Processor struct {
	Redis   *redis.Client
	Log     *log.Logger
	Store   *store.Repository
	Gateway string // Gateway URL, e.g., http://localhost:8090
}

func (p Processor) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.Log.Printf("worker loop stopped")
			return
		case <-ticker.C:
			p.drainOnce(ctx)
		}
	}
}

func (p Processor) drainOnce(ctx context.Context) {
	if p.Redis == nil {
		return
	}

	result, err := p.Redis.BLPop(ctx, 500*time.Millisecond, "wa:jobs").Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return
		}
		p.Log.Printf("redis pop error: %v", err)
		return
	}

	if len(result) < 2 {
		return
	}

	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		p.Log.Printf("invalid job payload: %v", err)
		return
	}

	// Handle different job types
	switch job.Type {
	case "message.send":
		p.handleMessageSend(ctx, job)
	default:
		p.Log.Printf("unknown job type: %s", job.Type)
	}
}

type messagePayload struct {
	MessageID   string `json:"messageId"`
	DeviceKey   string `json:"deviceKey"`
	ToNumber    string `json:"toNumber"`
	MessageBody string `json:"messageBody"`
}

type gatewayResponse struct {
	Status    string `json:"status"`
	MessageID string `json:"messageId"`
	Error     string `json:"error"`
}

func (p Processor) handleMessageSend(ctx context.Context, job Job) {
	var payload messagePayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		p.Log.Printf("failed to unmarshal message payload: %v", err)
		return
	}

	p.Log.Printf("sending message %s via device %s to %s", payload.MessageID, payload.DeviceKey, payload.ToNumber)

	// Call gateway to send message
	gatewayURL := fmt.Sprintf("%s/devices/%s/messages/send", p.Gateway, payload.DeviceKey)
	reqBody, _ := json.Marshal(map[string]string{
		"toNumber":    payload.ToNumber,
		"messageBody": payload.MessageBody,
	})

	resp, err := http.Post(gatewayURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		p.Log.Printf("failed to send message via gateway: %v", err)
		if p.Store != nil {
			_ = p.Store.UpdateMessageStatusByID(ctx, payload.MessageID, "failed", "", err.Error())
		}
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var gResp gatewayResponse
	if err := json.Unmarshal(respBody, &gResp); err != nil {
		p.Log.Printf("failed to parse gateway response: %v", err)
		if p.Store != nil {
			_ = p.Store.UpdateMessageStatusByID(ctx, payload.MessageID, "failed", "", "invalid gateway response")
		}
		return
	}

	if gResp.Status == "sent" {
		if p.Store != nil {
			_ = p.Store.UpdateMessageStatusByID(ctx, payload.MessageID, "sent", gResp.MessageID, "")
		}
		p.Log.Printf("message %s sent successfully (wa_id=%s)", payload.MessageID, gResp.MessageID)
	} else {
		if p.Store != nil {
			_ = p.Store.UpdateMessageStatusByID(ctx, payload.MessageID, "failed", "", gResp.Error)
		}
		p.Log.Printf("gateway returned error for message %s: %s", payload.MessageID, gResp.Error)
	}
}
