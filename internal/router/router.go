package router

import (
	"github.com/labstack/echo/v4"
)

// Router определяет интерфейс для регистрации маршрутов.
// Все роутеры приложения должны реализовывать этот интерфейс.
type Router interface {
	// RegisterRoutes регистрирует маршруты роутера в Echo instance.
	//
	// Параметры:
	//   - e: Echo instance для регистрации маршрутов
	//   - middlewares: вариативные параметры middleware (rate limiter, auth, cors, etc.)
	//
	// Каждая реализация Router решает, к каким маршрутам применять middleware,
	// используя Echo Groups для группировки маршрутов с общими middleware.
	RegisterRoutes(e *echo.Echo, middlewares ...echo.MiddlewareFunc)
}
