package router

import (
	"net/http"
	"url-shortener/internal/handler"

	"github.com/labstack/echo/v4"
)

type HomeRouter struct {
	handler *handler.HomeHandler
}

func NewHomeRouter(h *handler.HomeHandler) *HomeRouter {
	return &HomeRouter{handler: h}
}

func (r *HomeRouter) RegisterRoutes(e *echo.Echo, middlewares ...echo.MiddlewareFunc) {
	e.GET("/", r.handler.Home)

	e.GET("/health", r.healthCheck)

	e.RouteNotFound("/*", r.notFound)
}

func (r *HomeRouter) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
}

func (r *HomeRouter) notFound(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}
