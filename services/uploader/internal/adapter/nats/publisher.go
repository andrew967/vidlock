package nats

import (
	"fmt"

	"uploader/internal/config"

	nats "github.com/nats-io/nats.go"
)

type Publisher interface {
	Publish(subject string, data []byte, headers map[string]string) error
	EnsureStream(stream string) error
}

type jetStreamPublisher struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

func NewJetStreamPublisher(cfg *config.Config) (Publisher, error) {
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

	pub := &jetStreamPublisher{conn: conn, js: js}
	if err := pub.EnsureStream(cfg.NATS.Stream); err != nil {
		return nil, err
	}

	return pub, nil
}

func (p *jetStreamPublisher) EnsureStream(stream string) error {
	fmt.Println("✅ Ensuring stream VIDEO_UPLOADS with subjects: video.uploads.*, video.events")
	_, err := p.js.StreamInfo(stream)
	if err == nil {
		return nil
	}
	_, err = p.js.AddStream(&nats.StreamConfig{
		Name:     stream,
		Subjects: []string{"video.uploads.*", "video.events"},
		Storage:  nats.FileStorage,
	})

	return err
}

func (p *jetStreamPublisher) Publish(subject string, data []byte, headers map[string]string) error {
	fmt.Printf("→ PUBLISH to [%s], size: %d\n", subject, len(data))
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  nats.Header{},
	}
	for k, v := range headers {
		msg.Header.Set(k, v)
	}
	_, err := p.js.PublishMsg(msg)
	if err != nil {
		fmt.Printf("❌ Publish error: %v\n", err)
	}
	return err
}
