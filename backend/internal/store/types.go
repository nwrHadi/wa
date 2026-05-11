package store

import "time"

type User struct {
	ID           uint64    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Device struct {
	ID          uint64     `json:"id"`
	DeviceKey   string     `json:"deviceKey"`
	Label       string     `json:"label"`
	PhoneNumber string     `json:"phoneNumber,omitempty"`
	Status      string     `json:"status"`
	LastSeenAt  *time.Time `json:"lastSeenAt,omitempty"`
	QRText      string     `json:"qrText,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type Message struct {
	ID            uint64    `json:"id"`
	DeviceID      uint64    `json:"deviceId"`
	ExternalRefID string    `json:"externalRefId"`
	WAMessageID   string    `json:"waMessageId,omitempty"`
	ToNumber      string    `json:"toNumber"`
	Direction     string    `json:"direction"`
	Status        string    `json:"status"`
	Body          string    `json:"body"`
	ErrorMessage  string    `json:"errorMessage,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type Webhook struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	TargetURL    string    `json:"targetUrl"`
	Secret       string    `json:"secret,omitempty"`
	Enabled      bool      `json:"enabled"`
	EventFilters string    `json:"eventFilters,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type WebhookDelivery struct {
	ID           uint64     `json:"id"`
	WebhookID    uint64     `json:"webhookId"`
	EventID      string     `json:"eventId"`
	EventType    string     `json:"eventType"`
	Payload      string     `json:"payload"`
	Status       string     `json:"status"`
	AttemptCount int        `json:"attemptCount"`
	NextRetryAt  *time.Time `json:"nextRetryAt,omitempty"`
	LastError    string     `json:"lastError,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type DashboardSummary struct {
	ConnectedDevices int       `json:"connectedDevices"`
	Processing       int       `json:"processing"`
	Sent             int       `json:"sent"`
	Failed           int       `json:"failed"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type StreamEvent struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}
