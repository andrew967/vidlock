package handler

import (
	"net/http"

	"metadata/internal/config"
	"metadata/internal/domain"

	"github.com/gin-gonic/gin"
)

type VideoUseCase interface {
	GetVideoByID(id string) (*domain.Video, error)
	GetVideosByUser(userID string) ([]domain.Video, error)
}

type Handler struct {
	usecase VideoUseCase
	cfg     *config.Config
}

func NewHandler(usecase VideoUseCase, cfg *config.Config) *Handler {
	return &Handler{
		usecase: usecase,
		cfg:     cfg,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/videos/:id", h.GetVideo)

	authorized := router.Group("/my")
	authorized.Use(JWTMiddleware(h.cfg))
	{
		authorized.GET("/videos", h.GetMyVideos)
	}
}

func (h *Handler) GetVideo(c *gin.Context) {
	id := c.Param("id")
	video, err := h.usecase.GetVideoByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}
	c.JSON(http.StatusOK, video)
}

func (h *Handler) GetMyVideos(c *gin.Context) {
	userID := c.GetString("user_id")
	videos, err := h.usecase.GetVideosByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve videos"})
		return
	}
	c.JSON(http.StatusOK, videos)
}
