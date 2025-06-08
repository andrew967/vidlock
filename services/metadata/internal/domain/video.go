package domain

import "time"

type VideoStatus string

const (
	StatusPending VideoStatus = "pending"
	StatusReady   VideoStatus = "ready"
)

type Video struct {
	ID        string      `db:"id"`
	UserID    string      `db:"user_id"`
	FileName  string      `db:"file_name"`
	URL       string      `db:"url"`
	Status    VideoStatus `db:"status"`
	Size      int64       `db:"size"`
	CreatedAt time.Time   `db:"created_at"`
}
