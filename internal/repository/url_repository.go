package repository

import (
	"context"
	"url-shortener/internal/domain"

	"github.com/jmoiron/sqlx"
)

type UrlRepository struct {
	db *sqlx.DB
}

func NewUrlRepo(db *sqlx.DB) *UrlRepository {
	return &UrlRepository{db: db}
}

func (r *UrlRepository) Create(ctx context.Context, u *domain.Url) error {
	query := `
		INSERT INTO urls (long_url, code)
		VALUES (:long_url, :code)
		RETURNING id;
	`

	// можно еще использовать NamedQuery с последующим Next() -> Scan
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	defer func() { _ = stmt.Close() }()

	if err != nil {
		return err
	}
	if err = stmt.GetContext(ctx, u, u); err != nil {
		return err
	}

	return nil
}

func (r *UrlRepository) GetByCode(ctx context.Context, code string) (*domain.Url, error) {
	var u domain.Url

	query := `
		SELECT id, code, long_url, created_at, updated_at
		FROM urls
		WHERE code = $1
	`

	if err := r.db.GetContext(ctx, &u, query, code); err != nil {
		return nil, err
	}

	return &u, nil
}
