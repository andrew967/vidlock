package http

import (
	"fmt"
	"log"

	"uploader/internal/config"
	"uploader/internal/usecase/upload"

	"github.com/gin-gonic/gin"
)

func StartServer(cfg *config.Config, handler *upload.Handler) {
	r := gin.Default()

	r.POST("/upload", JWTMiddleware(cfg), handler.Upload)

	addr := fmt.Sprintf(":%d", cfg.HTTP.Port)
	log.Printf("Starting HTTP server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to run HTTP server: %v", err)
	}
}
