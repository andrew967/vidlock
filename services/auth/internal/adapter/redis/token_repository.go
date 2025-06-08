package redis

import (
	"auth/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type tokenRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewTokenRepository(client *redis.Client, ttl time.Duration) repository.TokenRepository {
	return &tokenRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *tokenRepository) StoreRefreshToken(ctx context.Context, userID string, token string) error {
	key := fmt.Sprintf("refresh:%s", userID)
	return r.client.Set(ctx, key, token, r.ttl).Err()
}

func (r *tokenRepository) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("refresh:%s", userID)
	result, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("refresh token not found")
	}
	return result, err
}

func (r *tokenRepository) DeleteRefreshToken(ctx context.Context, userID string) error {
	key := fmt.Sprintf("refresh:%s", userID)
	return r.client.Del(ctx, key).Err()
}
