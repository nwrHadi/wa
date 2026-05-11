package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users
		WHERE username = ?
	`, username)
	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *Repository) UpsertDevice(ctx context.Context, deviceKey, label, status, phoneNumber, qrText string, sessionBlob []byte, lastSeenAt *time.Time) (Device, error) {
	if strings.TrimSpace(deviceKey) == "" {
		return Device{}, fmt.Errorf("device key is required")
	}
	if strings.TrimSpace(label) == "" {
		label = deviceKey
	}

	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO devices (device_key, label, phone_number, status, last_seen_at)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			label = VALUES(label),
			phone_number = VALUES(phone_number),
			status = VALUES(status),
			last_seen_at = VALUES(last_seen_at),
			updated_at = CURRENT_TIMESTAMP
	`, deviceKey, label, phoneNumber, status, lastSeenAt)
	if err != nil {
		return Device{}, err
	}

	row := r.DB.QueryRowContext(ctx, `
		SELECT id, device_key, label, COALESCE(phone_number, ''), status, last_seen_at, created_at, updated_at
		FROM devices
		WHERE device_key = ?
	`, deviceKey)

	device, err := scanDevice(row)
	if err != nil {
		return Device{}, err
	}

	if sessionBlob != nil || qrText != "" {
		_, _ = r.DB.ExecContext(ctx, `
			INSERT INTO device_sessions (device_id, session_blob, qr_code_text, session_version)
			VALUES (?, ?, ?, 1)
			ON DUPLICATE KEY UPDATE
				session_blob = VALUES(session_blob),
				qr_code_text = VALUES(qr_code_text),
				updated_at = CURRENT_TIMESTAMP
		`, device.ID, sessionBlob, qrText)
		device.QRText = qrText
	}

	return device, nil
}

func (r *Repository) ListDevices(ctx context.Context) ([]Device, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT d.id, d.device_key, d.label, COALESCE(d.phone_number, ''), d.status, d.last_seen_at, d.created_at, d.updated_at,
		       COALESCE(ds.qr_code_text, '')
		FROM devices d
		LEFT JOIN device_sessions ds ON ds.device_id = d.id
		ORDER BY d.updated_at DESC, d.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]Device, 0)
	for rows.Next() {
		var device Device
		if err := rows.Scan(&device.ID, &device.DeviceKey, &device.Label, &device.PhoneNumber, &device.Status, &device.LastSeenAt, &device.CreatedAt, &device.UpdatedAt, &device.QRText); err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	return devices, rows.Err()
}

func (r *Repository) DeleteDeviceByKey(ctx context.Context, deviceKey string) (bool, error) {
	deviceKey = strings.TrimSpace(deviceKey)
	if deviceKey == "" {
		return false, fmt.Errorf("device key is required")
	}

	result, err := r.DB.ExecContext(ctx, `
		DELETE FROM devices
		WHERE device_key = ?
	`, deviceKey)
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (r *Repository) InsertMessage(ctx context.Context, deviceID uint64, externalRefID, toNumber, body, status string) (Message, error) {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO messages (device_id, external_ref_id, to_number, direction, status, body)
		VALUES (?, ?, ?, 'outbound', ?, ?)
	`, deviceID, externalRefID, toNumber, status, body)
	if err != nil {
		return Message{}, err
	}
	return r.MessageByExternalRef(ctx, externalRefID)
}

func (r *Repository) MessageByExternalRef(ctx context.Context, externalRefID string) (Message, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, device_id, external_ref_id, COALESCE(wa_message_id, ''), to_number, direction, status, body, COALESCE(error_message, ''), created_at, updated_at
		FROM messages
		WHERE external_ref_id = ?
	`, externalRefID)
	var m Message
	if err := row.Scan(&m.ID, &m.DeviceID, &m.ExternalRefID, &m.WAMessageID, &m.ToNumber, &m.Direction, &m.Status, &m.Body, &m.ErrorMessage, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return Message{}, err
	}
	return m, nil
}

func (r *Repository) UpdateMessageStatus(ctx context.Context, externalRefID, status, waMessageID, errorMessage string) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE messages
		SET status = ?, wa_message_id = COALESCE(NULLIF(?, ''), wa_message_id), error_message = ?, updated_at = CURRENT_TIMESTAMP
		WHERE external_ref_id = ?
	`, status, waMessageID, errorMessage, externalRefID)
	return err
}

// UpdateMessageStatusByID updates message status by message ID (string)
func (r *Repository) UpdateMessageStatusByID(ctx context.Context, messageID, status, waMessageID, errorMessage string) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE messages
		SET status = ?, wa_message_id = COALESCE(NULLIF(?, ''), wa_message_id), error_message = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, waMessageID, errorMessage, messageID)
	return err
}

func (r *Repository) InsertMessageEvent(ctx context.Context, messageID uint64, eventType string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO message_events (message_id, event_type, payload)
		VALUES (?, ?, ?)
	`, messageID, eventType, body)
	return err
}

func (r *Repository) ListMessages(ctx context.Context, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, device_id, external_ref_id, COALESCE(wa_message_id, ''), to_number, direction, status, body, COALESCE(error_message, ''), created_at, updated_at
		FROM messages
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Message, 0)
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.DeviceID, &m.ExternalRefID, &m.WAMessageID, &m.ToNumber, &m.Direction, &m.Status, &m.Body, &m.ErrorMessage, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (r *Repository) DashboardSummary(ctx context.Context) (DashboardSummary, error) {
	var s DashboardSummary
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices WHERE status = 'connected'`).Scan(&s.ConnectedDevices); err != nil {
		return s, err
	}
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM messages WHERE status = 'processing'`).Scan(&s.Processing); err != nil {
		return s, err
	}
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM messages WHERE status = 'sent'`).Scan(&s.Sent); err != nil {
		return s, err
	}
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM messages WHERE status = 'failed'`).Scan(&s.Failed); err != nil {
		return s, err
	}
	s.UpdatedAt = time.Now().UTC()
	return s, nil
}

func (r *Repository) CreateWebhook(ctx context.Context, name, targetURL, secret string, eventFilters string, enabled bool) (Webhook, error) {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO webhooks (name, target_url, secret, enabled, event_filters)
		VALUES (?, ?, ?, ?, ?)
	`, name, targetURL, secret, enabled, eventFilters)
	if err != nil {
		return Webhook{}, err
	}
	rows, err := r.ListWebhooks(ctx)
	if err != nil {
		return Webhook{}, err
	}
	if len(rows) == 0 {
		return Webhook{}, sql.ErrNoRows
	}
	return rows[0], nil
}

func (r *Repository) ListWebhooks(ctx context.Context) ([]Webhook, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, name, target_url, COALESCE(secret, ''), enabled, COALESCE(CAST(event_filters AS CHAR), ''), created_at, updated_at
		FROM webhooks
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Webhook, 0)
	for rows.Next() {
		var w Webhook
		var enabled int
		if err := rows.Scan(&w.ID, &w.Name, &w.TargetURL, &w.Secret, &enabled, &w.EventFilters, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		w.Enabled = enabled == 1
		items = append(items, w)
	}
	return items, rows.Err()
}

func (r *Repository) InsertWebhookDelivery(ctx context.Context, webhookID uint64, eventID, eventType, payload string, status string, attemptCount int, nextRetryAt *time.Time, lastError string) (WebhookDelivery, error) {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO webhook_deliveries (webhook_id, event_id, event_type, payload, status, attempt_count, next_retry_at, last_error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, webhookID, eventID, eventType, payload, status, attemptCount, nextRetryAt, lastError)
	if err != nil {
		return WebhookDelivery{}, err
	}
	return r.WebhookDeliveryByEvent(ctx, webhookID, eventID)
}

func (r *Repository) UpdateWebhookDelivery(ctx context.Context, id uint64, status string, attemptCount int, nextRetryAt *time.Time, lastError string) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE webhook_deliveries
		SET status = ?, attempt_count = ?, next_retry_at = ?, last_error = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, attemptCount, nextRetryAt, lastError, id)
	return err
}

func (r *Repository) WebhookDeliveryByEvent(ctx context.Context, webhookID uint64, eventID string) (WebhookDelivery, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, webhook_id, event_id, event_type, payload, status, attempt_count, next_retry_at, COALESCE(last_error, ''), created_at, updated_at
		FROM webhook_deliveries
		WHERE webhook_id = ? AND event_id = ?
	`, webhookID, eventID)
	var d WebhookDelivery
	if err := row.Scan(&d.ID, &d.WebhookID, &d.EventID, &d.EventType, &d.Payload, &d.Status, &d.AttemptCount, &d.NextRetryAt, &d.LastError, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return WebhookDelivery{}, err
	}
	return d, nil
}

func (r *Repository) DueWebhookDeliveries(ctx context.Context, limit int) ([]WebhookDelivery, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, webhook_id, event_id, event_type, payload, status, attempt_count, next_retry_at, COALESCE(last_error, ''), created_at, updated_at
		FROM webhook_deliveries
		WHERE status IN ('pending', 'retry')
		  AND (next_retry_at IS NULL OR next_retry_at <= CURRENT_TIMESTAMP)
		ORDER BY COALESCE(next_retry_at, created_at) ASC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]WebhookDelivery, 0)
	for rows.Next() {
		var d WebhookDelivery
		if err := rows.Scan(&d.ID, &d.WebhookID, &d.EventID, &d.EventType, &d.Payload, &d.Status, &d.AttemptCount, &d.NextRetryAt, &d.LastError, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *Repository) ListWebhookDeliveries(ctx context.Context, limit int) ([]WebhookDelivery, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, webhook_id, event_id, event_type, payload, status, attempt_count, next_retry_at, COALESCE(last_error, ''), created_at, updated_at
		FROM webhook_deliveries
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]WebhookDelivery, 0)
	for rows.Next() {
		var d WebhookDelivery
		if err := rows.Scan(&d.ID, &d.WebhookID, &d.EventID, &d.EventType, &d.Payload, &d.Status, &d.AttemptCount, &d.NextRetryAt, &d.LastError, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func scanDevice(row *sql.Row) (Device, error) {
	var device Device
	if err := row.Scan(&device.ID, &device.DeviceKey, &device.Label, &device.PhoneNumber, &device.Status, &device.LastSeenAt, &device.CreatedAt, &device.UpdatedAt); err != nil {
		return Device{}, err
	}
	return device, nil
}

// GetDevice retrieves a device by its key
func (r *Repository) GetDevice(ctx context.Context, deviceKey string) (*Device, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, device_key, label, COALESCE(phone_number, ''), status, last_seen_at, created_at, updated_at
		FROM devices
		WHERE device_key = ?
	`, deviceKey)

	device, err := scanDevice(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

// CreateMessage creates a new message record
func (r *Repository) CreateMessage(ctx context.Context, msg *Message) (*Message, error) {
	result, err := r.DB.ExecContext(ctx, `
		INSERT INTO messages (device_id, external_ref_id, direction, status, to_number, body, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, msg.DeviceID, msg.ExternalRefID, msg.Direction, msg.Status, msg.ToNumber, msg.Body, msg.CreatedAt, msg.UpdatedAt)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	msg.ID = uint64(id)
	return msg, nil
}

// GetMessage retrieves a message by ID
func (r *Repository) GetMessage(ctx context.Context, messageID string) (*Message, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, device_id, external_ref_id, COALESCE(wa_message_id, ''), to_number, direction, status, body, COALESCE(error_message, ''), created_at, updated_at
		FROM messages
		WHERE id = ?
	`, messageID)

	var m Message
	if err := row.Scan(&m.ID, &m.DeviceID, &m.ExternalRefID, &m.WAMessageID, &m.ToNumber, &m.Direction, &m.Status, &m.Body, &m.ErrorMessage, &m.CreatedAt, &m.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// GetMessageByExternalRef retrieves a message by external reference ID
func (r *Repository) GetMessageByExternalRef(ctx context.Context, externalRefID string) (*Message, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, device_id, external_ref_id, COALESCE(wa_message_id, ''), to_number, direction, status, body, COALESCE(error_message, ''), created_at, updated_at
		FROM messages
		WHERE external_ref_id = ?
	`, externalRefID)

	var m Message
	if err := row.Scan(&m.ID, &m.DeviceID, &m.ExternalRefID, &m.WAMessageID, &m.ToNumber, &m.Direction, &m.Status, &m.Body, &m.ErrorMessage, &m.CreatedAt, &m.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// ListMessages retrieves messages for a specific device with pagination
func (r *Repository) ListMessagesByDevice(ctx context.Context, deviceID string, limit, offset int) ([]Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, device_id, external_ref_id, COALESCE(wa_message_id, ''), to_number, direction, status, body, COALESCE(error_message, ''), created_at, updated_at
		FROM messages
		WHERE device_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, deviceID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Message, 0)
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.DeviceID, &m.ExternalRefID, &m.WAMessageID, &m.ToNumber, &m.Direction, &m.Status, &m.Body, &m.ErrorMessage, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}
