package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func OpenX(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxIdleConns(25)
	db.SetMaxOpenConns(5)
	db.SetConnMaxLifetime(5)

	return db, nil
}
