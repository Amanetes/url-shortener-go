package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
	"url-shortener/internal/domain"
	"url-shortener/internal/repository"

	"github.com/labstack/gommon/log"
	"github.com/redis/go-redis/v9"
)

const urlTTL = 10 * time.Minute

type UrlService struct {
	repo  *repository.UrlRepository
	cache *redis.Client
}

func NewUrlService(repo *repository.UrlRepository, rdb *redis.Client) *UrlService {
	return &UrlService{
		repo:  repo,
		cache: rdb,
	}
}

func (s *UrlService) Create(ctx context.Context, longUrl string) (*domain.Url, error) {
	code, err := generateShortCode()

	if err != nil {
		return nil, fmt.Errorf("failed to generate short code: %w", err)
	}

	u := &domain.Url{
		LongUrl: longUrl,
		Code:    code,
	}

	if err = s.cache.Set(ctx, "url:"+u.Code, u.LongUrl, urlTTL).Err(); err != nil {
		log.Errorf("failed to set url in cache: %v", err)
	}

	// Передаем контекст дальше в репозиторий
	if err = s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func generateShortCode() (string, error) {
	b := make([]byte, 8) // берем 64 бит рандома
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *UrlService) GetByCode(ctx context.Context, code string) (*domain.Url, error) {
	cacheKey := "url:" + code

	val, err := s.cache.Get(ctx, cacheKey).Result()

	if err == nil {
		// Нашли в кеше — возвращаем
		return &domain.Url{
			Code:    code,
			LongUrl: val,
		}, nil
	}

	// Редис лег - логируем
	if !errors.Is(err, redis.Nil) {
		log.Printf("redis error: %v", err)
	}

	u, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Нашли - кладем в кеш
	if err = s.cache.Set(ctx, cacheKey, u.LongUrl, urlTTL).Err(); err != nil {
		log.Printf("failed to set url in cache: %v", err)
	}

	return u, nil
}
