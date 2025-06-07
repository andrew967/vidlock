package nats

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"

	"processor/internal/config"
	"processor/internal/usecase"

	nats "github.com/nats-io/nats.go"
)

type Subscriber struct {
	js   nats.JetStreamContext
	conn *nats.Conn
}

func NewSubscriber(cfg *config.Config) (*Subscriber, error) {
	opts := []nats.Option{}
	if cfg.NATS.Token != "" {
		opts = append(opts, nats.Token(cfg.NATS.Token))
	}
	conn, err := nats.Connect(cfg.NATS.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream init: %w", err)
	}
	return &Subscriber{conn: conn, js: js}, nil
}

func (s *Subscriber) JetStream() nats.JetStreamContext {
	return s.js
}

func (s *Subscriber) SubscribeToEvents(processor usecase.ProcessorInterface) error {
	_, err := s.js.Subscribe("video.events", func(msg *nats.Msg) {
		videoID := msg.Header.Get("Video-ID")
		if videoID == "" {
			msg.Nak()
			return
		}

		fmt.Printf("📩 Event received: %s\n", videoID)

		go func() {
			err := processor.Process(context.Background(), videoID)
			if err != nil {
				fmt.Printf("❌ Processing error for %s: %v\n", videoID, err)
			}
			msg.Ack()
		}()
	}, nats.Durable("processor-durable"), nats.ManualAck())

	return err
}

type JetStreamFetcher struct {
	js nats.JetStreamContext
}

func NewChunkFetcher(js nats.JetStreamContext) *JetStreamFetcher {
	return &JetStreamFetcher{js: js}
}

func (f *JetStreamFetcher) FetchChunks(ctx context.Context, videoID string) (string, error) {
	subject := fmt.Sprintf("video.uploads.%s", videoID)
	consOpts := []nats.SubOpt{
		nats.DeliverAll(),
		nats.Durable(fmt.Sprintf("fetcher-%s", videoID)),
		nats.ManualAck(),
	}

	sub, err := f.js.PullSubscribe(subject, "", consOpts...)

	if err != nil {
		return "", fmt.Errorf("pull sub: %w", err)
	}

	chunks := make(map[int][]byte)
	total := 0
	for {
		msgs, err := sub.Fetch(10, nats.MaxWait(nats.DefaultTimeout))
		if err != nil {
			break // закончили
		}
		for _, msg := range msgs {
			chunkIdx, err := strconv.Atoi(msg.Header.Get("Chunk-Idx"))
			if err != nil {
				continue
			}
			chunks[chunkIdx] = msg.Data
			msg.Ack()
			total++
		}
	}

	var keys []int
	for k := range chunks {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	var buf bytes.Buffer
	for _, k := range keys {
		buf.Write(chunks[k])
	}

	tmpPath := fmt.Sprintf("/tmp/%s_raw.mp4", videoID)
	// tmpPath := fmt.Sprintf("%s/%s_raw.mp4", os.TempDir(), videoID)
	if err := writeTemp(tmpPath, buf.Bytes()); err != nil {
		return "", err
	}

	return tmpPath, nil
}

func writeTemp(path string, data []byte) error {
	return os.WriteFile(path, data, 0600)
}
