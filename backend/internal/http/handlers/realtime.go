package handlers

import (
	"fmt"
	"net/http"
	"time"

	"wa-platform/backend/internal/realtime"

	"github.com/labstack/echo/v4"
)

type RealtimeHandler struct {
	Hub *realtime.Hub
}

func (h RealtimeHandler) Stream(c echo.Context) error {
	if h.Hub == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "realtime unavailable"})
	}
	res := c.Response()
	res.Header().Set("Content-Type", "text/event-stream")
	res.Header().Set("Cache-Control", "no-cache")
	res.Header().Set("Connection", "keep-alive")
	res.WriteHeader(http.StatusOK)
	flusher, ok := res.Writer.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported")
	}

	updates := h.Hub.Subscribe()
	defer h.Hub.Unsubscribe(updates)

	ping := time.NewTicker(20 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case msg, ok := <-updates:
			if !ok {
				return nil
			}
			_, _ = fmt.Fprintf(res.Writer, "data: %s\n\n", msg)
			flusher.Flush()
		case <-ping.C:
			_, _ = fmt.Fprint(res.Writer, ": ping\n\n")
			flusher.Flush()
		}
	}
}
