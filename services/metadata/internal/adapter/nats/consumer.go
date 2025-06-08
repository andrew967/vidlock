package nats

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"metadata/internal/domain"

	"github.com/nats-io/nats.go"
)

type VideoRepository interface {
	Create(ctx context.Context, v *domain.Video) error
	UpdateStatusAndURL(ctx context.Context, id string, status domain.VideoStatus, url string) error
}

type Consumer struct {
	js       nats.JetStreamContext
	repo     VideoRepository
	subjects []string
}

func NewConsumer(js nats.JetStreamContext, repo VideoRepository) *Consumer {
	return &Consumer{
		js:   js,
		repo: repo,
		subjects: []string{
			"video.events",
			"video.processed.*",
		},
	}
}

func (c *Consumer) Start() error {
	for _, subj := range c.subjects {
		_, err := c.js.Subscribe(subj, c.handleMessage)
		if err != nil {
			return err
		}
		log.Println("Subscribed to", subj)
	}
	return nil
}

func (c *Consumer) handleMessage(msg *nats.Msg) {
	ctx := context.Background()

	switch {
	case msg.Subject == "video.events":
		videoID := string(msg.Data)
		userID := msg.Header.Get("User-ID")
		fileName := msg.Header.Get("File-Name")

		if userID == "" || fileName == "" {
			log.Println("❌ Missing User-ID or File-Name in video.events headers")
			return
		}

		video := &domain.Video{
			ID:        videoID,
			UserID:    userID,
			FileName:  fileName,
			Status:    domain.StatusPending,
			Size:      0, // при необходимости можешь добавить Video-Size из header
			CreatedAt: time.Now(),
		}

		if err := c.repo.Create(ctx, video); err != nil {
			log.Println("❌ Error saving video:", err)
			return
		}
		log.Println("✅ Video created from headers:", video.ID)

	case strings.HasPrefix(msg.Subject, "video.processed."):
		var payload struct {
			VideoID string `json:"video_id"`
			URL     string `json:"url"`
		}
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Println("Error decoding video.processed.*:", err)
			return
		}

		if err := c.repo.UpdateStatusAndURL(ctx, payload.VideoID, domain.StatusReady, payload.URL); err != nil {
			log.Println("Error updating video:", err)
			return
		}
		log.Println("Video updated to ready:", payload.VideoID)
	}
}
