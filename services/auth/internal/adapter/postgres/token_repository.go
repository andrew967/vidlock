package postgres

import (
	"context"
	"fmt"

	"auth/internal/repository"
)

type tokenRepository struct {
	db DBTX
}

func NewTokenRepository(db DBTX) repository.TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) StoreRefreshToken(ctx context.Context, userID string, token string) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token
	`
	_, err := r.db.ExecContext(ctx, query, userID, token)
	return err
}

func (r *tokenRepository) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	query := `
		SELECT token FROM refresh_tokens WHERE user_id = $1
	`
	var token string
	err := r.db.QueryRowxContext(ctx, query, userID).Scan(&token)
	if err != nil {
		return "", fmt.Errorf("refresh token not found: %w", err)
	}
	return token, nil
}

func (r *tokenRepository) DeleteRefreshToken(ctx context.Context, userID string) error {
	query := `
		DELETE FROM refresh_tokens WHERE user_id = $1
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
