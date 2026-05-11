package service

import (
	"context"
	"encoding/json"
	"time"
)

type MessageQueuePayload struct {
	DeviceKey     string `json:"deviceKey"`
	ExternalRefID string `json:"externalRefId"`
	ToNumber      string `json:"toNumber"`
	Body          string `json:"body"`
}

func (q *Queue) EnqueueMessage(ctx context.Context, payload MessageQueuePayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return q.Enqueue(ctx, QueueJob{Type: MessageJobType, Payload: body, CreatedAt: time.Now().UTC()})
}
