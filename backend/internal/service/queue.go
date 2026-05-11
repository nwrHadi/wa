package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	MessageJobType      = "message.send"
	WebhookDeliveryType = "webhook.delivery"
	QueueKey            = "wa:jobs"
	DeadLetterQueueKey  = "wa:dlq"
)

type Queue struct {
	Redis *redis.Client
}

type QueueJob struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"createdAt"`
}

func NewQueue(redisClient *redis.Client) *Queue {
	return &Queue{Redis: redisClient}
}

func (q *Queue) Enabled() bool {
	return q != nil && q.Redis != nil
}

func (q *Queue) Enqueue(ctx context.Context, job QueueJob) error {
	if !q.Enabled() {
		return errors.New("queue unavailable")
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now().UTC()
	}
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.Redis.RPush(ctx, QueueKey, body).Err()
}

func (q *Queue) EnqueueDeadLetter(ctx context.Context, reason string, job QueueJob) error {
	if !q.Enabled() {
		return errors.New("queue unavailable")
	}
	deadLetter := map[string]any{
		"reason":    reason,
		"job":       job,
		"createdAt": time.Now().UTC(),
	}
	body, err := json.Marshal(deadLetter)
	if err != nil {
		return err
	}
	return q.Redis.RPush(ctx, DeadLetterQueueKey, body).Err()
}

func (j QueueJob) String() string {
	return fmt.Sprintf("type=%s", j.Type)
}
