package upload

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"uploader/internal/adapter/nats"
	"uploader/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	cfg       *config.Config
	publisher nats.Publisher
	bufPool   sync.Pool
}

func NewHandler(cfg *config.Config, publisher nats.Publisher) *Handler {
	return &Handler{
		cfg:       cfg,
		publisher: publisher,
		bufPool: sync.Pool{
			New: func() any {
				return make([]byte, cfg.App.ChunkSize)
			},
		},
	}
}

func (h *Handler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file"})
		return
	}
	defer file.Close()

	videoID := uuid.New().String()
	subject := fmt.Sprintf("video.uploads.%s", videoID)

	h.publisher.Publish("video.events", []byte(videoID), map[string]string{
		"Video-ID":  videoID,
		"File-Name": header.Filename,
		"Subject":   subject,
	})

	idx := 0
	for {
		buf := h.bufPool.Get().([]byte)
		n, err := file.Read(buf)
		if n > 0 {
			h.publisher.Publish(subject, buf[:n], map[string]string{
				"Video-ID":  videoID,
				"Chunk-Idx": fmt.Sprintf("%d", idx),
			})
			idx++
		}
		h.bufPool.Put(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "read error"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"video_id":    videoID,
		"chunks_sent": idx,
	})
}
