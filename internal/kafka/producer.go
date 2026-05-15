package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
	kafka "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(broker),
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: 5 * time.Second,
			RequiredAcks: kafka.RequireOne,
		},
	}
}

func (p *Producer) Publish(topic string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(context.Background(), kafka.Message{
		Topic: topic,
		Value: data,
	})
	if err != nil {
		log.Error().Err(err).Str("topic", topic).Msg("failed to publish kafka message")
		return err
	}

	log.Debug().Str("topic", topic).Msg("kafka message published")
	return nil
}

func (p *Producer) Close() {
	p.writer.Close()
}
