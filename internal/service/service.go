package service

import (
	"url-shortener/internal/repository"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	UrlService *UrlService
}

// New инициализирует все сервисы, прокидывая в них нужные репозитории
func New(repos *repository.Repository, rdb *redis.Client) *Service {
	return &Service{
		UrlService: NewUrlService(repos.Url, rdb),
	}
}
