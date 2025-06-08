package usecase

import (
	"auth/internal/entity"
	"context"
)

type AuthUseCase interface {
	Register(ctx context.Context, email string, password string) error
	Login(ctx context.Context, email string, password string) (accessToken string, refreshToken string, err error)
	Refresh(ctx context.Context, userID string, oldRefreshToken string) (accessToken string, refreshToken string, err error)
	Logout(ctx context.Context, userID string) error
	DeleteAccount(ctx context.Context, userID string) error
	GetMe(ctx context.Context, userID string) (*entity.User, error)
}
