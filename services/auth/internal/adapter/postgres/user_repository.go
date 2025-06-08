package postgres

import (
	"auth/internal/entity"
	"auth/internal/repository"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db DBTX
}

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...any) error
}

func NewUserRepository(db DBTX) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, hashed_password, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.HashedPassword, user.CreatedAt)
	return err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, hashed_password, created_at
		FROM users
		WHERE email = $1
	`
	var user entity.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, email, hashed_password, created_at
		FROM users
		WHERE id = $1
	`
	var user entity.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (r *userRepository) DeleteByID(ctx context.Context, id string) error {
	query := `
		DELETE FROM users WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
