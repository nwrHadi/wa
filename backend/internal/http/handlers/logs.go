package handlers

import (
	"net/http"

	"wa-platform/backend/internal/store"

	"github.com/labstack/echo/v4"
)

type LogsHandler struct {
	Store *store.Repository
}

func (h LogsHandler) ListMessages(c echo.Context) error {
	if h.Store != nil {
		items, err := h.Store.ListMessages(c.Request().Context(), 100)
		if err == nil {
			return c.JSON(http.StatusOK, map[string]interface{}{"items": items})
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"items": []map[string]interface{}{},
	})
}
