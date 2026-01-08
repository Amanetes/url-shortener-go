package router

import (
	"url-shortener/internal/handler"

	"github.com/labstack/echo/v4"
)

type URLRouter struct {
	handler *handler.URLHandler
}

func NewURLRouter(h *handler.URLHandler) *URLRouter {
	return &URLRouter{handler: h}
}

func (r *URLRouter) RegisterRoutes(e *echo.Echo, middlewares ...echo.MiddlewareFunc) {
	protected := e.Group("", middlewares...)
	protected.POST("/shorten", r.handler.Shorten)

	e.GET("/:code", r.handler.Redirect)
}
