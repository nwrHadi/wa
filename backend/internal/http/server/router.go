package server

import (
	"wa-platform/backend/internal/config"
	"wa-platform/backend/internal/http/handlers"
	appmw "wa-platform/backend/internal/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewRouter(cfg config.Config, h Handlers) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("internal.shared.token", cfg.InternalSharedToken)
			return next(c)
		}
	})
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())

	e.GET("/healthz", h.Health.Healthz)
	e.GET("/readyz", h.Health.Readyz)
	e.POST("/api/v1/auth/login", h.Auth.Login)
	e.POST("/internal/events/baileys", h.InternalEvents.FromBaileys)
	e.GET("/api/v1/realtime", h.Realtime.Stream)

	api := e.Group("/api/v1", appmw.JWT(cfg.JWTSecret))
	registerProtectedRoutes(api, h)

	// Support Cloudflare path-based routing via /wa without breaking existing clients.
	wa := e.Group("/wa")
	wa.GET("/healthz", h.Health.Healthz)
	wa.GET("/readyz", h.Health.Readyz)
	wa.POST("/api/v1/auth/login", h.Auth.Login)
	wa.POST("/internal/events/baileys", h.InternalEvents.FromBaileys)
	wa.GET("/api/v1/realtime", h.Realtime.Stream)

	waAPI := wa.Group("/api/v1", appmw.JWT(cfg.JWTSecret))
	registerProtectedRoutes(waAPI, h)

	return e
}

func registerProtectedRoutes(api *echo.Group, h Handlers) {
	api.GET("/dashboard/summary", h.Dashboard.Summary)
	api.GET("/dashboard/timeline", h.Dashboard.Timeline)
	api.GET("/devices", h.Devices.List)
	api.POST("/devices", h.Devices.Create)
	api.POST("/devices/:deviceKey/connect", h.Messages.DeviceConnect)
	api.POST("/devices/:deviceKey/disconnect", h.Messages.DeviceDisconnect)
	api.GET("/devices/:deviceKey/status", h.Messages.DeviceStatus)
	api.DELETE("/devices", h.Devices.Delete)
	api.DELETE("/devices/:deviceKey", h.Devices.Delete)
	api.GET("/logs/messages", h.Logs.ListMessages)
	api.POST("/messages/send", h.Messages.Send)
	api.GET("/messages/:messageId", h.Messages.Get)
	api.GET("/devices/:deviceId/messages", h.Messages.ListByDevice)
	api.GET("/webhooks", h.Webhooks.List)
}

type Handlers struct {
	Health         handlers.HealthHandler
	Auth           handlers.AuthHandler
	Dashboard      handlers.DashboardHandler
	Devices        handlers.DevicesHandler
	Messages       handlers.MessagesHandler
	Logs           handlers.LogsHandler
	Webhooks       handlers.WebhookHandler
	InternalEvents handlers.InternalEventsHandler
	Realtime       handlers.RealtimeHandler
}
