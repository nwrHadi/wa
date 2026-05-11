package handlers

import (
	"net/http"

	"wa-platform/backend/internal/store"

	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	Store *store.Repository
}

func (h DashboardHandler) Summary(c echo.Context) error {
	if h.Store != nil {
		summary, err := h.Store.DashboardSummary(c.Request().Context())
		if err == nil {
			return c.JSON(http.StatusOK, summary)
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"connectedDevices": 0,
		"processing":       0,
		"sent":             0,
		"failed":           0,
		"updatedAt":        "",
	})
}

func (h DashboardHandler) Timeline(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"points": []map[string]interface{}{},
	})
}
