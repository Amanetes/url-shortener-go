package router

import (
	"net/http"
	"url-shortener/internal/handler"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, h *handler.Handler, limiterMiddleware echo.MiddlewareFunc) {
	// root
	e.GET("/", h.Home, limiterMiddleware)

	// health
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
	})

	// urls
	e.POST("/shorten", h.Shorten, limiterMiddleware)
	e.GET("/:code", h.Redirect)

	e.RouteNotFound("/*", func(c echo.Context) error { return c.NoContent(http.StatusNotFound) })
}
