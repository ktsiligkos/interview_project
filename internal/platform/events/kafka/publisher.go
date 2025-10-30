package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"

	companyservice "github.com/ktsiligkos/xm_project/internal/service/company"
)

// Publisher writes company events to a Kafka topic
type Publisher struct {
	topic string
	w     *kafka.Writer
}

// NewPublisher constructs a Publisher
func NewPublisher(brokers []string, topic string) *Publisher {
	return &Publisher{
		topic: topic,
		w: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireOne,
			BatchTimeout: 10 * time.Millisecond,
		},
	}
}

// PublishCompanyEvent serialises the event and sends it to Kafka
func (p *Publisher) PublishCompanyEvent(ctx context.Context, event companyservice.CompanyEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(event.Company.ID),
		Value: payload,
		Time:  time.Now(),
	}

	return p.w.WriteMessages(ctx, msg)
}

// Safely close the publisher
func (p *Publisher) Close() error {
	if p == nil || p.w == nil {
		return nil
	}
	return p.w.Close()
}
