package messaging

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/notification-service/config"
	"github.com/segmentio/kafka-go"
)

func StartConsumers(ctx context.Context) {
	for _, topic := range Topics {
		go startTopicConsumer(ctx, topic, config.Cfg.Kafka.GroupIDs.Notification, func(ctx context.Context, msg kafka.Message) {
			handleNotificaion(msg)
		})
	}
}

func startTopicConsumer(ctx context.Context, topic, groupID string, handler func(ctx context.Context, msg kafka.Message)) {
	if config.Cfg.Kafka.Brokers == "" {
		log.Fatal("KAFKA_BROKERS environment variable is not set")
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
	case OrderCreatedTopic:
		notificationMessage = "Order has been created."
	case OrderUpdatedTopic:
		notificationMessage = "Order has been updated."
	case PaymentSucceededTopic:
		notificationMessage = "Payment was successfull."
	case PaymentFailedTopic:
		notificationMessage = "Payment failed."
	case OrderPickedUpTopic:
		notificationMessage = "Order picked up by courier."
	case OrderDeliveredTopic:
		notificationMessage = "Order delivered."
	default:
		notificationMessage = "An unknown event occured."
	}

	log.Printf("[NOTIFICATION] For Order ID: %s -> %s", orderID, notificationMessage)
}
