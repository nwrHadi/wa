package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"wa-platform/backend/internal/service"

	"github.com/labstack/echo/v4"
)

type InternalEventsHandler struct {
	Events *service.EventService
}

type inboundEventEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func (h InternalEventsHandler) FromBaileys(c echo.Context) error {
	parts := strings.Fields(c.Request().Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing internal token"})
	}

	expected, _ := c.Get("internal.shared.token").(string)
	if expected == "" || parts[1] != expected {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid internal token"})
	}

	if h.Events == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "event service unavailable"})
	}

	var envelope inboundEventEnvelope
	if err := c.Bind(&envelope); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	if envelope.Type == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "event type is required"})
	}
	if err := h.Events.Handle(c.Request().Context(), service.EventEnvelope{Type: envelope.Type, Payload: envelope.Payload}); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusAccepted, map[string]string{"status": "received"})
}
