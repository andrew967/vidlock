package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pressly/goose/v3"

	"metadata/internal/adapter/nats"
	"metadata/internal/adapter/postgres"
	"metadata/internal/config"
	"metadata/internal/handler"
	infra "metadata/internal/infrastructure"
	"metadata/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	db, err := infra.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}
	if err := goose.Up(db.DB, "./migrations"); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	nc, js, err := infra.NewJetStream(cfg)
	if err != nil {
		log.Fatalf("nats error: %v", err)
	}
	defer nc.Drain()

	videoRepo := postgres.NewVideoRepository(db)
	videoUC := usecase.NewVideoUseCase(videoRepo)
	consumer := nats.NewConsumer(js, videoRepo)
	if err := consumer.Start(); err != nil {
		log.Fatalf("consumer error: %v", err)
	}

	router := gin.Default()
	h := handler.NewHandler(videoUC, cfg)
	h.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("🚀 metadata-service listening on :%d", cfg.HTTP.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
