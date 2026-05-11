package handlers

import (
	"encoding/json"
	"net/http"

	"wa-platform/backend/internal/service"
	"wa-platform/backend/internal/store"

	"github.com/labstack/echo/v4"
)

type WebhookHandler struct {
	Store *store.Repository
	Queue *service.Queue
}

type createWebhookRequest struct {
	Name         string `json:"name"`
	TargetURL    string `json:"targetUrl"`
	Secret       string `json:"secret"`
	EventFilters string `json:"eventFilters"`
	Enabled      bool   `json:"enabled"`
}

func (h WebhookHandler) List(c echo.Context) error {
	if h.Store != nil {
		items, err := h.Store.ListWebhooks(c.Request().Context())
		if err == nil {
			return c.JSON(http.StatusOK, map[string]interface{}{"items": items})
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"items": []map[string]interface{}{},
	})
}

func (h WebhookHandler) Create(c echo.Context) error {
	var req createWebhookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	if h.Store == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "store unavailable"})
	}
	webhook, err := h.Store.CreateWebhook(c.Request().Context(), req.Name, req.TargetURL, req.Secret, req.EventFilters, req.Enabled)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if h.Queue != nil && h.Queue.Enabled() {
		payload, _ := json.Marshal(webhook)
		_ = h.Queue.Enqueue(c.Request().Context(), service.QueueJob{Type: "webhook.created", Payload: payload})
	}
	return c.JSON(http.StatusCreated, webhook)
}
