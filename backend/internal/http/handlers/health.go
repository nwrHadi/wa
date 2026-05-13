package handlers

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	DB            *sql.DB
	Redis         *redis.Client
	RedisOptional bool
}

func (h HealthHandler) Healthz(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h HealthHandler) Readyz(c echo.Context) error {
	if err := h.DB.Ping(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "db_unavailable"})
	}
	if h.Redis != nil {
		if err := h.Redis.Ping(c.Request().Context()).Err(); err != nil {
			if h.RedisOptional {
				return c.JSON(http.StatusOK, map[string]string{"status": "ready", "redis": "optional_unavailable"})
			}
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "redis_unavailable"})
		}
	} else if !h.RedisOptional {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "redis_unavailable"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
}
