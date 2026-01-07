package main

import (
	"url-shortener/internal/config"
	"url-shortener/internal/db"
	"url-shortener/internal/handler"
	"url-shortener/internal/repository"
	"url-shortener/internal/server"
	"url-shortener/internal/service"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func main() {
	l, _ := zap.NewProduction()
	defer func() {
		_ = l.Sync()
	}()

	cfg, err := config.Get()
	if err != nil {
		l.Fatal("failed to load config", zap.Error(err))
	}

	pg, err := db.OpenX(cfg.Db.Dsn())
	if err != nil {
		l.Fatal("failed to connect to database", zap.Error(err))
	}

	defer func() {
		_ = pg.Close()
	}()

	redisClient, err := db.NewRedis(cfg.Redis)
	if err != nil {
		l.Fatal("failed to connect to redis", zap.Error(err))
	}

	defer func() {
		_ = redisClient.Close()
	}()

	validate := validator.New()

	urlRepo := repository.NewUrlRepo(pg)
	urlService := service.NewURLService(urlRepo, redisClient, l)
	homeHandler := handler.NewHomeHandler()
	urlHandler := handler.NewURLHandler(urlService, cfg, validate)

	s := server.NewServer(cfg, homeHandler, urlHandler, l)
	l.Info("Listening on port", zap.String("port", cfg.Server.Port))

	s.Run()
}
