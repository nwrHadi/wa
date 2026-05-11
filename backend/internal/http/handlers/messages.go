package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"wa-platform/backend/internal/realtime"
	"wa-platform/backend/internal/service"
	"wa-platform/backend/internal/store"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MessagesHandler struct {
	Store      *store.Repository
	Queue      *service.Queue
	Hub        *realtime.Hub
	GatewayURL string
}

type sendMessageRequest struct {
	DeviceKey      string `json:"deviceKey"`
	ToNumber       string `json:"toNumber"`
	MessageBody    string `json:"messageBody"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type sendMessageResponse struct {
	MessageID     uint64    `json:"messageId"`
	ExternalRefID string    `json:"externalRefId"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

type gatewaySendResponse struct {
	Status    string `json:"status"`
	MessageID string `json:"messageId"`
	Error     string `json:"error"`
}

type gatewayDeviceStatusResponse struct {
	IsConnected bool   `json:"isConnected"`
	QRCode      string `json:"qrCode"`
	PhoneNumber string `json:"phoneNumber"`
}

// Send message from a device to WhatsApp recipient
func (h MessagesHandler) Send(c echo.Context) error {
	var req sendMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	// Validate inputs
	req.DeviceKey = strings.TrimSpace(req.DeviceKey)
	req.ToNumber = strings.TrimSpace(req.ToNumber)
	req.MessageBody = strings.TrimSpace(req.MessageBody)

	if req.DeviceKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "deviceKey is required"})
	}
	if req.ToNumber == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "toNumber is required"})
	}
	if req.MessageBody == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "messageBody is required"})
	}

	if h.Store == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "store unavailable"})
	}

	ctx := c.Request().Context()

	// Check device exists
	device, err := h.Store.GetDevice(ctx, req.DeviceKey)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "device not found"})
	}
	if device == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "device not found"})
	}

	gatewayStatus, err := h.getGatewayDeviceStatus(ctx, req.DeviceKey)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": fmt.Sprintf("failed to verify gateway status: %v", err)})
	}

	if !gatewayStatus.IsConnected {
		_, _ = h.Store.UpsertDevice(ctx, device.DeviceKey, device.Label, "disconnected", device.PhoneNumber, "", nil, nil)
		return c.JSON(http.StatusConflict, map[string]string{"error": "device not connected"})
	}

	phoneNumber := device.PhoneNumber
	if gatewayStatus.PhoneNumber != "" {
		phoneNumber = gatewayStatus.PhoneNumber
	}
	if device.Status != "connected" || phoneNumber != device.PhoneNumber {
		updatedDevice, updateErr := h.Store.UpsertDevice(ctx, device.DeviceKey, device.Label, "connected", phoneNumber, "", nil, nil)
		if updateErr == nil {
			device = &updatedDevice
		}
	}

	// Check idempotency
	externalRefID := req.IdempotencyKey
	if externalRefID == "" {
		externalRefID = uuid.New().String()
	}

	existingMsg, err := h.Store.GetMessageByExternalRef(ctx, externalRefID)
	if err == nil && existingMsg != nil {
		// Already sent, return existing message
		return c.JSON(http.StatusOK, map[string]interface{}{
			"messageId":     existingMsg.ID,
			"externalRefId": externalRefID,
			"status":        existingMsg.Status,
			"createdAt":     existingMsg.CreatedAt,
		})
	}

	// Create message record
	message := &store.Message{
		DeviceID:      device.ID,
		ExternalRefID: externalRefID,
		Direction:     "outbound",
		Status:        "processing",
		ToNumber:      req.ToNumber,
		Body:          req.MessageBody,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	msg, err := h.Store.CreateMessage(ctx, message)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create message"})
	}

	// Publish event to realtime hub
	if h.Hub != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"messageId": msg.ID,
			"deviceKey": device.DeviceKey,
			"toNumber":  req.ToNumber,
			"status":    "processing",
			"createdAt": msg.CreatedAt,
		})
		h.Hub.Publish(realtime.Event{Type: "message.created", Payload: json.RawMessage(payload)})
	}

	waMessageID, sendErr := h.sendViaGateway(ctx, device.DeviceKey, req.ToNumber, req.MessageBody)
	if sendErr != nil {
		sendErrText := sendErr.Error()
		statusCode := http.StatusBadGateway
		if strings.Contains(strings.ToLower(sendErrText), "device not connected") {
			statusCode = http.StatusConflict
			_, _ = h.Store.UpsertDevice(ctx, device.DeviceKey, device.Label, "disconnected", device.PhoneNumber, "", nil, nil)
		}

		_ = h.Store.UpdateMessageStatusByID(ctx, fmt.Sprintf("%d", msg.ID), "failed", "", sendErr.Error())
		_ = h.Store.InsertMessageEvent(ctx, msg.ID, "message.failed", map[string]interface{}{
			"deviceKey":     device.DeviceKey,
			"externalRefId": externalRefID,
			"toNumber":      req.ToNumber,
			"body":          req.MessageBody,
			"status":        "failed",
			"errorMessage":  sendErr.Error(),
		})

		if h.Hub != nil {
			payload, _ := json.Marshal(map[string]interface{}{
				"messageId":     msg.ID,
				"externalRefId": externalRefID,
				"deviceKey":     device.DeviceKey,
				"toNumber":      req.ToNumber,
				"status":        "failed",
				"errorMessage":  sendErr.Error(),
				"createdAt":     msg.CreatedAt,
			})
			h.Hub.Publish(realtime.Event{Type: "message.failed", Payload: json.RawMessage(payload)})
		}

		return c.JSON(statusCode, map[string]interface{}{
			"messageId":     msg.ID,
			"externalRefId": externalRefID,
			"status":        "failed",
			"error":         sendErrText,
			"createdAt":     msg.CreatedAt,
		})
	}

	if err := h.Store.UpdateMessageStatusByID(ctx, fmt.Sprintf("%d", msg.ID), "sent", waMessageID, ""); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update message status"})
	}
	if err := h.Store.InsertMessageEvent(ctx, msg.ID, "message.sent", map[string]interface{}{
		"deviceKey":     device.DeviceKey,
		"externalRefId": externalRefID,
		"waMessageId":   waMessageID,
		"toNumber":      req.ToNumber,
		"body":          req.MessageBody,
		"status":        "sent",
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to store message event"})
	}

	if h.Hub != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"messageId":     msg.ID,
			"externalRefId": externalRefID,
			"waMessageId":   waMessageID,
			"deviceKey":     device.DeviceKey,
			"toNumber":      req.ToNumber,
			"status":        "sent",
			"createdAt":     msg.CreatedAt,
		})
		h.Hub.Publish(realtime.Event{Type: "message.sent", Payload: json.RawMessage(payload)})
	}

	return c.JSON(http.StatusOK, sendMessageResponse{
		MessageID:     msg.ID,
		ExternalRefID: externalRefID,
		Status:        "sent",
		CreatedAt:     msg.CreatedAt,
	})
}

func (h MessagesHandler) sendViaGateway(ctx context.Context, deviceKey, toNumber, messageBody string) (string, error) {
	gatewayURL := strings.TrimSpace(h.GatewayURL)
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8090"
	}

	reqBody, err := json.Marshal(map[string]string{
		"toNumber":    toNumber,
		"messageBody": messageBody,
	})
	if err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/devices/%s/messages/send", strings.TrimRight(gatewayURL, "/"), deviceKey), bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	responseBody, _ := io.ReadAll(response.Body)
	var gatewayResp gatewaySendResponse
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &gatewayResp); err != nil {
			return "", fmt.Errorf("invalid gateway response")
		}
	}

	if response.StatusCode >= http.StatusBadRequest {
		if gatewayResp.Error != "" {
			return "", fmt.Errorf(gatewayResp.Error)
		}
		return "", fmt.Errorf("gateway send failed with status %d", response.StatusCode)
	}
	if gatewayResp.Status != "sent" {
		if gatewayResp.Error != "" {
			return "", fmt.Errorf(gatewayResp.Error)
		}
		return "", fmt.Errorf("gateway did not confirm delivery")
	}

	return gatewayResp.MessageID, nil
}

func (h MessagesHandler) getGatewayDeviceStatus(ctx context.Context, deviceKey string) (*gatewayDeviceStatusResponse, error) {
	gatewayURL := strings.TrimSpace(h.GatewayURL)
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8090"
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/devices/%s/status", strings.TrimRight(gatewayURL, "/"), deviceKey), nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("gateway status failed with status %d", response.StatusCode)
	}

	var status gatewayDeviceStatusResponse
	if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

// Get message details
func (h MessagesHandler) Get(c echo.Context) error {
	messageID := c.Param("messageId")

	if h.Store == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "store unavailable"})
	}

	msg, err := h.Store.GetMessage(c.Request().Context(), messageID)
	if err != nil || msg == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "message not found"})
	}

	return c.JSON(http.StatusOK, msg)
}

// List messages for a device
func (h MessagesHandler) ListByDevice(c echo.Context) error {
	deviceID := c.Param("deviceId")
	limit := 50
	offset := 0

	if l := c.QueryParam("limit"); l != "" {
		if parsed := parseInt(l); parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if parsed := parseInt(o); parsed >= 0 {
			offset = parsed
		}
	}

	if h.Store == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "store unavailable"})
	}

	messages, err := h.Store.ListMessagesByDevice(c.Request().Context(), deviceID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list messages"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"items":  messages,
		"count":  len(messages),
		"limit":  limit,
		"offset": offset,
	})
}

func parseInt(s string) int {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0
	}
	return n
}
