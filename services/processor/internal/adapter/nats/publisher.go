package nats

import (
	"encoding/json"
	"fmt"

	natsgo "github.com/nats-io/nats.go"
)

type EventPublisher struct {
	js     natsgo.JetStreamContext
	stream string
}

func NewEventPublisher(js natsgo.JetStreamContext, stream string) *EventPublisher {
	return &EventPublisher{js: js, stream: stream}
}

func (p *EventPublisher) EnsureStream() error {
	info, err := p.js.StreamInfo(p.stream)
	if err != nil {

		if err == natsgo.ErrStreamNotFound {
			_, err := p.js.AddStream(&natsgo.StreamConfig{
				Name:     p.stream,
				Subjects: []string{"video.uploads.*", "video.events", "video.processed.*", "video.progress.*"},
				Storage:  natsgo.FileStorage,
			})
			return err
		}
		return err
	}

	subjectSet := make(map[string]bool)
	for _, s := range info.Config.Subjects {
		subjectSet[s] = true
	}

	changed := false
	for _, required := range []string{"video.uploads.*", "video.events", "video.processed.*", "video.progress.*"} {
		if !subjectSet[required] {
			info.Config.Subjects = append(info.Config.Subjects, required)
			changed = true
		}
	}

	if changed {
		_, err = p.js.UpdateStream(&info.Config)
		if err != nil {
			return fmt.Errorf("update stream subjects: %w", err)
		}
	}

	return nil
}

func (p *EventPublisher) PublishProcessed(videoID string, url string) error {
	subject := fmt.Sprintf("video.processed.%s", videoID)
	data, _ := json.Marshal(map[string]interface{}{
		"video_id": videoID,
		"status":   "processed",
		"url":      url,
	})
	_, err := p.js.Publish(subject, data)
	return err
}

func (p *EventPublisher) PublishProgress(videoID string, percent int) error {
	subject := fmt.Sprintf("video.progress.%s", videoID)
	data, _ := json.Marshal(map[string]interface{}{
		"video_id": videoID,
		"progress": percent,
	})
	_, err := p.js.Publish(subject, data)
	return err
}
