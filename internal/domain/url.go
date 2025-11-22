package domain

import (
	"time"
)

type Url struct {
	ID        int        `db:"id"`
	LongUrl   string     `db:"long_url"`
	Code      string     `db:"code"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}
