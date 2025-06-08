package usecase

import (
	"context"

	"metadata/internal/domain"
)

type VideoRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Video, error)
	FindByUser(ctx context.Context, userID string) ([]domain.Video, error)
}

type VideoUseCase struct {
	repo VideoRepository
}

func NewVideoUseCase(repo VideoRepository) *VideoUseCase {
	return &VideoUseCase{
		repo: repo,
	}
}

func (uc *VideoUseCase) GetVideoByID(id string) (*domain.Video, error) {
	return uc.repo.FindByID(context.Background(), id)
}

func (uc *VideoUseCase) GetVideosByUser(userID string) ([]domain.Video, error) {
	return uc.repo.FindByUser(context.Background(), userID)
}
