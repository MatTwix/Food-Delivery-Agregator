package messaging

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/payments-service/config"
	"github.com/segmentio/kafka-go"
)

type OrderCreatedEvent struct {
	ID         string  `json:"id"`
	TotalPrice float64 `json:"total_price"`
}

func StartConsumers(ctx context.Context, p *Producer) {
	go startTopicConsumer(ctx, OrderCreatedTopic, config.Cfg.Kafka.GroupIDs.Orders, func(ctx context.Context, msg kafka.Message) {
		handleOrderCreated(ctx, msg, p)
	})
}

func startTopicConsumer(ctx context.Context, topic, groupID string, handler func(ctx context.Context, msg kafka.Message)) {
	if config.Cfg.Kafka.Brokers == "" {
		slog.Error("KAFKA_BROKERS environment variable is not set")
		os.Exit(1)
	}

	brokers := strings.Split(config.Cfg.Kafka.Brokers, ",")

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: 1 * time.Second,
		StartOffset:    kafka.LastOffset,
	})

	slog.Info("starting Kafka consumer", "topic", topic, "group_id", groupID)

	defer r.Close()

	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping consumer due to context cancellation", "topic", topic)
			return
		default:
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					slog.Info("context cancelled, stopping consumer", "topic", topic)
					return
				}
				slog.Error("failed to read message", "topic", topic, "error", err)
				continue
			}
			slog.Info("processing message", "topic", topic)
			handler(ctx, m)

			if err := r.CommitMessages(ctx, m); err != nil {
				slog.Error("failed to commit message offset", "error", err)
			}
		}
	}
}

func handleOrderCreated(ctx context.Context, msg kafka.Message, p *Producer) {
	var event OrderCreatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.Error("failed to unmarshal event", "event", OrderCreatedTopic, "error", err)
		return
	}

	slog.Info("processing payment", "order_od", event.ID, "total_price", event.TotalPrice)

	// imitating payment process

	time.Sleep(5 * time.Second)

	if rand.Float32() < 0.8 {
		slog.Info("payment SUCCEEDED", "order_id", event.ID)
		p.Produce(ctx, PaymentSucceededTopic, []byte(event.ID), msg.Value)
	} else {
		slog.Info("payment FAILED", "order_id", event.ID)
		p.Produce(ctx, PaymentFailedTopic, []byte(event.ID), msg.Value)
	}
}
