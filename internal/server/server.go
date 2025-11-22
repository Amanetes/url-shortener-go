package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/handler"
	"url-shortener/internal/router"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

type Server struct {
	cfg *config.Config
	*echo.Echo
	*handler.Handler
}

func NewServer(cfg *config.Config, h *handler.Handler) *Server {
	e := echo.New()

	if cfg.App.Env == "development" {
		e.Debug = true
	}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Pre(middleware.RemoveTrailingSlash())

	e.Validator = handler.NewValidator()

	router.SetupRoutes(e, h, IPRateLimiter())

	return &Server{
		cfg:  cfg,
		Echo: e,
	}
}

func (s *Server) Run() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		addr := fmt.Sprintf(":%s", s.cfg.Server.Port)

		if err := s.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.Logger.Fatal("Shutting down the server")
		}
	}()

	<-ctx.Done()
	s.Logger.Printf("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownGracePeriod)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		s.Logger.Fatal(err)
	}
}

func IPRateLimiter() echo.MiddlewareFunc {
	rate := limiter.Rate{
		Period: 1 * time.Second,
		Limit:  10,
	}

	store := memory.NewStore()
	l := limiter.New(store, rate)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			ip := c.RealIP()
			// key := c.Request().Header.Get("X-API-Key")

			limiterCtx, err := l.Get(c.Request().Context(), ip)
			if err != nil {
				log.Warnf("IPRateLimit - ipRateLimiter.Get - err: %v, %s on %s", err, ip, c.Request().URL)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"success": false,
					"message": err,
				})
			}

			h := c.Response().Header()
			h.Set("X-RateLimit-Limit", strconv.FormatInt(limiterCtx.Limit, 10))
			h.Set("X-RateLimit-Remaining", strconv.FormatInt(limiterCtx.Remaining, 10))
			h.Set("X-RateLimit-Reset", strconv.FormatInt(limiterCtx.Reset, 10))

			if limiterCtx.Reached {
				log.Printf("Too Many Requests from %s on %s", ip, c.Request().URL)
				return c.JSON(http.StatusTooManyRequests, echo.Map{
					"success": false,
					"message": "Too Many Requests on " + c.Request().URL.String(),
				})
			}

			// log.Printf("%s request continue", c.RealIP())
			return next(c)
		}
	}
}
