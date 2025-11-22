package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	Url *UrlRepository
}

func New(db *sqlx.DB) *Repository {
	return &Repository{
		Url: NewUrlRepo(db),
	}
}
