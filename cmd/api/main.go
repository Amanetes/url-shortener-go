package main

import (
	"url-shortener/internal/config"
	"url-shortener/internal/db"
	"url-shortener/internal/handler"
	"url-shortener/internal/repository"
	"url-shortener/internal/server"
	"url-shortener/internal/service"

	"github.com/labstack/gommon/log"
)

func main() {
	cfg, err := config.Get()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.OpenX(cfg.Db.Dsn())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	defer func() {
		_ = database.Close()
	}()

	redisClient, err := db.NewRedis(cfg.Redis)
	if err != nil {
		log.Warnf("failed to init redis: %v", err)
		return
	}

	repos := repository.New(database)
	services := service.New(repos, redisClient)
	handlers := handler.New(services, cfg)

	s := server.NewServer(cfg, handlers)
	s.Logger.Printf("Starting application in %s mode", cfg.App.Env)
	s.Logger.Printf("Listening on port %s", cfg.Server.Port)

	s.Run()
}
