package infra

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"metadata/internal/config"
)

func NewPostgres(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}
	return db, nil
}
