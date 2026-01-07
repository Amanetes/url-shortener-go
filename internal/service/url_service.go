package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"url-shortener/internal/domain"
	repositoryInterface "url-shortener/internal/repository/interface"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const urlTTL = 10 * time.Minute

type URLService struct {
	repo  repositoryInterface.URLRepository
	cache *redis.Client
	l     *zap.Logger
}

func NewURLService(repo repositoryInterface.URLRepository, cache *redis.Client, l *zap.Logger) *URLService {
	return &URLService{repo, cache, l}
}

func (s *URLService) Create(ctx context.Context, longUrl string) (*domain.Url, error) {
	if err := validateURL(longUrl); err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	code, err := generateShortCode()

	if err != nil {
		return nil, fmt.Errorf("failed to generate short code: %w", err)
	}

	u := &domain.Url{
		LongUrl: longUrl,
		Code:    code,
	}

	if err = s.cache.Set(ctx, "url:"+u.Code, u.LongUrl, urlTTL).Err(); err != nil {
		s.l.Error("failed to set url in cache", zap.Error(err))
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

func validateURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return errors.New("url cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("url scheme must be http or https")
	}

	if parsedURL.Host == "" {
		return errors.New("url must have a valid host")
	}

	return nil
}

func (s *URLService) GetByCode(ctx context.Context, code string) (*domain.Url, error) {
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
		s.l.Warn("failed to get url from cache", zap.Error(err))
	}

	u, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Нашли - кладем в кеш
	if err = s.cache.Set(ctx, cacheKey, u.LongUrl, urlTTL).Err(); err != nil {
		s.l.Error("failed to set url in cache", zap.Error(err))
	}

	return u, nil
}
