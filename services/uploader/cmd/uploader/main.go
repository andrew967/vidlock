package main

import (
	"log"

	"uploader/internal/adapter/http"
	"uploader/internal/adapter/nats"
	"uploader/internal/config"
	"uploader/internal/usecase/upload"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	publisher, err := nats.NewJetStreamPublisher(cfg)
	if err != nil {
		log.Fatalf("nats publisher error: %v", err)
	}

	handler := upload.NewHandler(cfg, publisher)

	http.StartServer(cfg, handler)
}
