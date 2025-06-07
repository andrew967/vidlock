package usecase

import (
	"context"
	"fmt"
	"log"
	"os"
)

type VideoAssembler interface {
	Assemble(chunks [][]byte) ([]byte, error)
}

type KeyStore interface {
	Save(videoID, chunkID string, key []byte) error
}

type ChunkFetcher interface {
	FetchChunks(ctx context.Context, videoID string) (string /*path to assembled raw video*/, error)
}

type WatermarkProcessor interface {
	ApplyWatermark(inputPath string) (string /*path to watermarked video*/, error)
}

type ChunkSplitter interface {
	Split(inputPath string) ([]string /*paths to chunk files*/, error)
}

type ChunkEncryptor interface {
	Encrypt(filePath string) (encryptedPath string, key []byte, err error)
}

type IPFSUploader interface {
	Upload(ctx context.Context, filePath string) (string /*ipfs URL*/, error)
}

type EventPublisher interface {
	PublishProcessed(videoID string, url string) error
	PublishProgress(videoID string, percent int) error
}

type ProcessorInterface interface {
	Process(ctx context.Context, videoID string) error
}
type Processor struct {
	fetcher     ChunkFetcher
	assembler   VideoAssembler
	watermarker WatermarkProcessor
	splitter    ChunkSplitter
	encryptor   ChunkEncryptor
	keyStore    KeyStore
	ipfs        IPFSUploader
	publisher   EventPublisher
}

func NewProcessor(
	f ChunkFetcher,
	a VideoAssembler,
	w WatermarkProcessor,
	s ChunkSplitter,
	e ChunkEncryptor,
	k KeyStore,
	ip IPFSUploader,
	pub EventPublisher,
) ProcessorInterface {
	return &Processor{
		fetcher:     f,
		assembler:   a,
		watermarker: w,
		splitter:    s,
		encryptor:   e,
		keyStore:    k,
		ipfs:        ip,
		publisher:   pub,
	}
}

func deleteIfExists(path string) {
	_ = os.Remove(path)
}

func (p *Processor) Process(ctx context.Context, videoID string) error {
	rawPath, err := p.fetcher.FetchChunks(ctx, videoID)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	defer deleteIfExists(rawPath)

	watermarkedPath, err := p.watermarker.ApplyWatermark(rawPath)
	if err != nil {
		return fmt.Errorf("watermark: %w", err)
	}
	defer deleteIfExists(watermarkedPath)

	chunkPaths, err := p.splitter.Split(watermarkedPath)
	if err != nil {
		return fmt.Errorf("split: %w", err)
	}
	for _, p := range chunkPaths {
		defer deleteIfExists(p)
	}

	total := len(chunkPaths)
	for i, chunkPath := range chunkPaths {
		encPath, key, err := p.encryptor.Encrypt(chunkPath)
		if err != nil {
			log.Printf("encrypt error: %v", err)
			continue
		}
		defer deleteIfExists(encPath)

		chunkID := fmt.Sprintf("%s_%03d", videoID, i)
		_ = p.keyStore.Save(videoID, chunkID, key)

		url, err := p.ipfs.Upload(ctx, encPath)
		if err != nil {
			log.Printf("upload failed: %v", err)
			continue
		}

		p.publisher.PublishProgress(videoID, (i+1)*100/total)
		log.Printf("Uploaded %s to IPFS: %s", chunkPath, url)
	}

	return p.publisher.PublishProcessed(videoID, fmt.Sprintf("ipfs://video/%s", videoID))
}
