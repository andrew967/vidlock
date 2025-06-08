package postgres

import (
	"context"

	"metadata/internal/domain"

	"github.com/jmoiron/sqlx"
)

type VideoRepository struct {
	db *sqlx.DB
}

func NewVideoRepository(db *sqlx.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

func (r *VideoRepository) Create(ctx context.Context, v *domain.Video) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO videos (id, user_id, file_name, url, status, size, created_at)
		VALUES (:id, :user_id, :file_name, :url, :status, :size, :created_at)
	`, v)
	return err
}

func (r *VideoRepository) UpdateStatusAndURL(ctx context.Context, id string, status domain.VideoStatus, url string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE videos SET status = $1, url = $2 WHERE id = $3
	`, status, url, id)
	return err
}

func (r *VideoRepository) FindByID(ctx context.Context, id string) (*domain.Video, error) {
	var v domain.Video
	err := r.db.GetContext(ctx, &v, `SELECT * FROM videos WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VideoRepository) FindByUser(ctx context.Context, userID string) ([]domain.Video, error) {
	var videos []domain.Video
	err := r.db.SelectContext(ctx, &videos, `SELECT * FROM videos WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	return videos, err
}
