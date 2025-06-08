package usecase

import (
	"auth/config"
	"auth/internal/entity"
	"auth/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authUseCase struct {
	cfg       *config.Config
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
}

func NewAuthUseCase(cfg *config.Config, userRepo repository.UserRepository, tokenRepo repository.TokenRepository) AuthUseCase {
	return &authUseCase{
		cfg:       cfg,
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

func (a *authUseCase) Register(ctx context.Context, email string, password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &entity.User{
		ID:             uuid.NewString(),
		Email:          email,
		HashedPassword: string(hashed),
		CreatedAt:      time.Now(),
	}

	return a.userRepo.Create(ctx, user)
}

func (a *authUseCase) Login(ctx context.Context, email string, password string) (string, string, error) {
	user, err := a.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := a.generateJWT(user.ID, a.cfg.JWT.AccessTokenTTL)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := a.generateJWT(user.ID, a.cfg.JWT.RefreshTokenTTL)
	if err != nil {
		return "", "", err
	}

	if err := a.tokenRepo.StoreRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (a *authUseCase) Refresh(ctx context.Context, userID string, oldRefreshToken string) (string, string, error) {
	stored, err := a.tokenRepo.GetRefreshToken(ctx, userID)
	if err != nil || stored != oldRefreshToken {
		return "", "", errors.New("unauthorized")
	}

	accessToken, err := a.generateJWT(userID, a.cfg.JWT.AccessTokenTTL)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := a.generateJWT(userID, a.cfg.JWT.RefreshTokenTTL)
	if err != nil {
		return "", "", err
	}

	if err := a.tokenRepo.StoreRefreshToken(ctx, userID, refreshToken); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (a *authUseCase) Logout(ctx context.Context, userID string) error {
	return a.tokenRepo.DeleteRefreshToken(ctx, userID)
}

func (a *authUseCase) DeleteAccount(ctx context.Context, userID string) error {
	return a.userRepo.DeleteByID(ctx, userID)
}

func (a *authUseCase) GetMe(ctx context.Context, userID string) (*entity.User, error) {
	return a.userRepo.FindByID(ctx, userID)
}

func (a *authUseCase) generateJWT(userID string, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(ttl).Unix(),
	})
	return token.SignedString([]byte(a.cfg.JWT.SecretKey))
}
