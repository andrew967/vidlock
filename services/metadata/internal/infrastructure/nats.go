package infra

import (
	"fmt"

	natsgo "github.com/nats-io/nats.go"

	"metadata/internal/config"
)

func NewJetStream(cfg *config.Config) (*natsgo.Conn, natsgo.JetStreamContext, error) {
	opts := []natsgo.Option{}
	if cfg.NATS.Token != "" {
		opts = append(opts, natsgo.Token(cfg.NATS.Token))
	}

	conn, err := natsgo.Connect(cfg.NATS.URL, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		return nil, nil, fmt.Errorf("jetstream init: %w", err)
	}

	return conn, js, nil
}
