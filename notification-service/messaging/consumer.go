package messaging

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/notification-service/config"
	"github.com/segmentio/kafka-go"
)

const (
	GroupID = "notification-service-group"
)

func StartConsumers(ctx context.Context) {
	for _, topic := range Topics {
		go startTopicConsumer(ctx, topic, GroupID, func(ctx context.Context, msg kafka.Message) {
			handleNotificaion(msg)
		})
	}
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

func handleNotificaion(msg kafka.Message) {
	orderID := string(msg.Key)

	var notificationMessage string

	switch msg.Topic {
	case string(OrderCreatedTopic):
		notificationMessage = "Order has been created."
	case string(OrderUpdatedTopic):
		notificationMessage = "Order has been updated."
	case string(PaymentSucceededTopic):
		notificationMessage = "Payment was successfull."
	case string(PaymentFailedTopic):
		notificationMessage = "Payment failed."
	case string(OrderPickedUpTopic):
		notificationMessage = "Order picked up by courier."
	case string(OrderDeliveredTopic):
		notificationMessage = "Order delivered."
	default:
		notificationMessage = "An unknown event occured."
	}

	log.Printf("[NOTIFICATION] For Order ID: %s -> %s", orderID, notificationMessage)
}
