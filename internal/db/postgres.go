package db

import (
	"fmt"
	"url-shortener/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func OpenX(cfg config.DbConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.Dsn())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return db, nil
}
