package router

import (
	"net/http"
	"url-shortener/internal/handler"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(
	e *echo.Echo,
	home *handler.HomeHandler,
	url *handler.URLHandler,
	limiterMiddleware echo.MiddlewareFunc,
) {
	// root
	e.GET("/", home.Home, limiterMiddleware)

	// health
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
	})

	// URLs
	e.POST("/shorten", url.Shorten, limiterMiddleware)
	e.GET("/:code", url.Redirect)

	e.RouteNotFound("/*", func(c echo.Context) error { return c.NoContent(http.StatusNotFound) })
}
