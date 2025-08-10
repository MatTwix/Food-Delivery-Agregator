package messaging

import (
	"context"
	"encoding/json"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/payments-service/config"
	"github.com/segmentio/kafka-go"
)

type OrderCreatedEvent struct {
	ID         string  `json:"id"`
	TotalPrice float64 `json:"total_price"`
}

const (
	GroupID = "payments-service-group"
)

func StartConsumers(ctx context.Context, p *Producer) {
	go startTopicConsumer(ctx, OrderCreatedTopic, GroupID, func(ctx context.Context, msg kafka.Message) {
		handleOrderCreated(ctx, msg, p)
	})
}

func startTopicConsumer(ctx context.Context, topic Topic, groupID string, handler func(ctx context.Context, msg kafka.Message)) {
	cfg := config.LoadConfig()

	if cfg.KafkaBrokers == "" {
		log.Fatal("KAFKA_BROKERS environment variable is not set")
	}

	brokers := strings.Split(cfg.KafkaBrokers, ",")

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          string(topic),
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: 1 * time.Second,
		StartOffset:    kafka.LastOffset,
	})

	log.Printf("Starting Kafka consumer for topic '%s' with group ID '%s'", topic, groupID)

	defer r.Close()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping consumer for topic '%s' due to context cancellation", topic)
			return
		default:
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Printf("Context cancelled, stopping consumer for topic '%s'", topic)
					return
				}
				log.Printf("Error reading message from topic '%s': %v", topic, err)
				continue
			}
			log.Printf("Processing message from topic '%s'", topic)
			handler(ctx, m)

			if err := r.CommitMessages(ctx, m); err != nil {
				log.Printf("Error committing message offset: %v", err)
			}
		}
	}
}

func handleOrderCreated(ctx context.Context, msg kafka.Message, p *Producer) {
	var event OrderCreatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("Error unmarshaling order.created event: %v", err)
		return
	}

	log.Printf("Processing payment for order %s with total price %.2f", event.ID, event.TotalPrice)

	// imitating payment process

	time.Sleep(5 * time.Second)

	if rand.Float32() < 0.8 {
		log.Printf("Payment for order %s SUCCEEDED", event.ID)
		p.Produce(ctx, PaymentSucceededTopic, []byte(event.ID), msg.Value)
	} else {
		log.Printf("Payment for order %s FAILED", event.ID)
		p.Produce(ctx, PaymentFailedTopic, []byte(event.ID), msg.Value)
	}
}
