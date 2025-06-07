package main

import (
	"log"

	"processor/internal/adapter/crypto"
	"processor/internal/adapter/ffmpeg"
	"processor/internal/adapter/ipfs"
	"processor/internal/adapter/nats"
	"processor/internal/adapter/vault"
	"processor/internal/config"
	"processor/internal/usecase"
)

func main() {
	log.Println("🚀 Processor starting...")

	cfg := config.Load()
	if err := config.LoadVaultSecrets(cfg); err != nil {
		log.Fatalf("🔒 Vault load error: %v", err)
	}

	natsSub, err := nats.NewSubscriber(cfg)
	if err != nil {
		log.Fatalf("🔌 NATS connect error: %v", err)
	}
	js := natsSub.JetStream()

	fetcher := nats.NewChunkFetcher(js)
	watermarker := ffmpeg.NewWatermarkProcessor("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf")
	splitter := ffmpeg.NewChunkSplitter(10)
	encryptor := crypto.NewChunkEncryptor()
	keyStore, err := vault.NewVaultKeyStore(cfg.Vault.Address, cfg.Vault.Token, "videos")
	if err != nil {
		log.Fatalf("🔐 Vault keystore error: %v", err)
	}
	uploader := ipfs.NewIPFSUploader(cfg.IPFS.APIAddress)
	publisher := nats.NewEventPublisher(js, cfg.NATS.Stream)
	if err != nil {
		log.Fatalf("failed to init publisher: %v", err)
	}
	if err = publisher.EnsureStream(); err != nil {
		log.Fatalf("failed to ensure stream: %v", err)
	} else {
		log.Println("✅ Stream ensured successfully")
	}
	processor := usecase.NewProcessor(
		fetcher,
		nil,
		watermarker,
		splitter,
		encryptor,
		keyStore,
		uploader,
		publisher,
	)

	if err := natsSub.SubscribeToEvents(processor); err != nil {
		log.Fatalf("📡 Subscribe error: %v", err)
	}

	log.Println("✅ Processor is listening to video.events...")

	select {}
}
