package main

import (
	"auth/config"
	"auth/internal/adapter/postgres"
	redisrepo "auth/internal/adapter/redis"
	"auth/internal/handler/http"
	"auth/internal/usecase"
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	goose "github.com/pressly/goose/v3"
)

func runMigrations(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("goose db open error: %w", err)
	}
	defer db.Close()

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose up error: %w", err)
	}

	return nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	if err := runMigrations(cfg.Postgres.DSN); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	pg, err := postgres.New(cfg)
	if err != nil {
		log.Fatalf("PostgreSQL init error: %v", err)
	}

	redisClient, err := redisrepo.NewClient(cfg)
	if err != nil {
		log.Fatalf("Redis init error: %v", err)
	}

	userRepo := postgres.NewUserRepository(pg.Conn)
	tokenRepo := redisrepo.NewTokenRepository(redisClient, cfg.JWT.RefreshTokenTTL)

	authUC := usecase.NewAuthUseCase(cfg, userRepo, tokenRepo)

	router := gin.Default()
	http.NewHandler(router, authUC, cfg)

	addr := fmt.Sprintf(":%d", cfg.HTTP.Port)
	log.Printf("Auth Service listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
