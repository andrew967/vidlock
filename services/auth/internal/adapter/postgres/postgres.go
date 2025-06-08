package postgres

import (
	"fmt"

	"auth/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DB struct {
	Conn *sqlx.DB
}

func New(cfg *config.Config) (*DB, error) {
	dsn := cfg.Postgres.DSN
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("PostgreSQL not responding: %w", err)
	}

	return &DB{Conn: db}, nil
}
