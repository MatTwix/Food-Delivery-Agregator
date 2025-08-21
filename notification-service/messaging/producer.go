package messaging

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/MatTwix/Food-Delivery-Agregator/notification-service/config"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer() (*Producer, error) {
	if config.Cfg.Kafka.Brokers == "" {
		slog.Error("KAFKA_BROKERS environment variable is not set")
		os.Exit(1)
	}

	brokers := strings.Split(config.Cfg.Kafka.Brokers, ",")

	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}

	return &Producer{writer: w}, nil
}

func (p *Producer) Produce(ctx context.Context, topic string, key, value []byte) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	})
	if err != nil {
		slog.Error("failed to write message to Kafka topic", "topic", topic, "error", err)
		return err
	}

	slog.Info("message sent to topic", "topic", topic)

	return nil
}

func (p *Producer) Close() {
	if err := p.writer.Close(); err != nil {
		slog.Error("failed to close Kafka writer", "error", err)
	}
}
