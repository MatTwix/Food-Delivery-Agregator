package messaging

import (
	"context"
	"log"
	"strings"

	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/config"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer() (*Producer, error) {
	if config.Cfg.Kafka.Broker == "" {
		log.Fatal("KAFKA_BROKERS environment variable is not set")
	}

	brokers := strings.Split(config.Cfg.Kafka.Broker, ",")

	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}

	return &Producer{writer: w}, nil
}

func (p *Producer) Produce(ctx context.Context, topic string, key, value []byte) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: string(topic),
		Key:   key,
		Value: value,
	})
	if err != nil {
		log.Printf("Failed to write message to Kafka topic %s: %v", topic, err)
		return err
	}

	log.Printf("Message sent to topic %s", topic)

	return nil
}

func (p *Producer) Close() {
	if err := p.writer.Close(); err != nil {
		log.Printf("Error closing Kafka writer: %v", err)
	}
}
