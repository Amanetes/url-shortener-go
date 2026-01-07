package repositoryInterface

import (
	"context"
	"url-shortener/internal/domain"
)

type URLRepository interface {
	Create(ctx context.Context, u *domain.Url) error
	GetByCode(ctx context.Context, code string) (*domain.Url, error)
}
