package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"wa-platform/backend/internal/realtime"
	"wa-platform/backend/internal/service"
	"wa-platform/backend/internal/store"

	"github.com/labstack/echo/v4"
)

type DevicesHandler struct {
	Store *store.Repository
	Queue *service.Queue
	Hub   *realtime.Hub
}

type createDeviceRequest struct {
	DeviceKey   string `json:"deviceKey"`
	Label       string `json:"label"`
	Status      string `json:"status"`
	PhoneNumber string `json:"phoneNumber"`
	QRText      string `json:"qrText"`
}

type deleteDeviceRequest struct {
	DeviceKey string `json:"deviceKey"`
}

func (h DevicesHandler) List(c echo.Context) error {
	if h.Store != nil {
		items, err := h.Store.ListDevices(c.Request().Context())
		if err == nil {
			return c.JSON(http.StatusOK, map[string]any{"items": items})
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"items": []map[string]interface{}{},
	})
}

func (h DevicesHandler) Create(c echo.Context) error {
	var req createDeviceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	req.DeviceKey = strings.TrimSpace(req.DeviceKey)
	if req.DeviceKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "deviceKey is required"})
	}
	if req.Label == "" {
		req.Label = req.DeviceKey
	}
	if req.Status == "" {
		req.Status = "disconnected"
	}
	if h.Store == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "store unavailable"})
	}
	device, err := h.Store.UpsertDevice(c.Request().Context(), req.DeviceKey, req.Label, req.Status, req.PhoneNumber, req.QRText, nil, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if h.Hub != nil {
		payload, _ := json.Marshal(device)
		h.Hub.Publish(realtime.Event{Type: "device.created", Payload: json.RawMessage(payload)})
	}
	return c.JSON(http.StatusCreated, device)
}

func (h DevicesHandler) Delete(c echo.Context) error {
	deviceKey := strings.TrimSpace(c.Param("deviceKey"))
	if deviceKey == "" {
		deviceKey = strings.TrimSpace(c.QueryParam("deviceKey"))
	}
	if deviceKey == "" {
		var req deleteDeviceRequest
		if err := c.Bind(&req); err == nil {
			deviceKey = strings.TrimSpace(req.DeviceKey)
		}
	}
	if deviceKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "deviceKey is required"})
	}
	if h.Store == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "store unavailable"})
	}

	deleted, err := h.Store.DeleteDeviceByKey(c.Request().Context(), deviceKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if !deleted {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "device not found"})
	}

	if h.Hub != nil {
		payload, _ := json.Marshal(map[string]string{"deviceKey": deviceKey})
		h.Hub.Publish(realtime.Event{Type: "device.deleted", Payload: json.RawMessage(payload)})
	}

	return c.NoContent(http.StatusNoContent)
}
