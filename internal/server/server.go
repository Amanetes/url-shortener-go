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
	"url-shortener/internal/router"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.uber.org/zap"
)

type Server struct {
	cfg *config.Config
	*echo.Echo
	l *zap.Logger
}

func NewServer(
	cfg *config.Config,
	routers []router.Router,
	l *zap.Logger,
) *Server {
	e := echo.New()

	if cfg.App.Env == "development" {
		e.Debug = true
	}

	// Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:           true,
		LogMethod:        true,
		LogStatus:        true,
		LogLatency:       true,
		LogRemoteIP:      true,
		LogUserAgent:     true,
		LogContentLength: true,
		LogResponseSize:  true,
		LogError:         true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			fields := []zap.Field{
				zap.String("method", v.Method),
				zap.String("uri", v.URI),
				zap.Int("status", v.Status),
				zap.Duration("latency", v.Latency),
				zap.String("remote_ip", v.RemoteIP),
				zap.String("user_agent", v.UserAgent),
				zap.String("bytes_in", v.ContentLength),
				zap.Int64("bytes_out", v.ResponseSize),
			}

			if v.Error != nil {
				fields = append(fields, zap.String("error", v.Error.Error()))
			}

			l.Info("http_request", fields...)
			return nil
		},
	}))
	e.Use(middleware.Recover())

	e.Pre(middleware.RemoveTrailingSlash())

	rateLimiter := IPRateLimiter(l)

	for _, r := range routers {
		r.RegisterRoutes(e, rateLimiter)
	}

	return &Server{
		cfg:  cfg,
		Echo: e,
		l:    l,
	}
}

func (s *Server) Run() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		addr := fmt.Sprintf(":%s", s.cfg.Server.Port)

		if err := s.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.l.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	s.l.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownGracePeriod)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		s.l.Fatal("Failed to shutdown server gracefully", zap.Error(err))
	}
}

func IPRateLimiter(logger *zap.Logger) echo.MiddlewareFunc {
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
				logger.Warn("IPRateLimit - ipRateLimiter.Get error",
					zap.Error(err),
					zap.String("ip", ip),
					zap.String("url", c.Request().URL.String()),
				)
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
				logger.Info("Too Many Requests",
					zap.String("ip", ip),
					zap.String("url", c.Request().URL.String()),
				)
				return c.JSON(http.StatusTooManyRequests, echo.Map{
					"success": false,
					"message": "Too Many Requests on " + c.Request().URL.String(),
				})
			}

			// logger.Debug("request continue", zap.String("ip", c.RealIP()))
			return next(c)
		}
	}
}
