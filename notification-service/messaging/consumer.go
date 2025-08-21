package messaging

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/notification-service/config"
	"github.com/segmentio/kafka-go"
)

type NotifiableEvent struct {
	UserID string `json:"user_id"`
}

type NotificationEvent struct {
	UserID  string `json:"user_id"`
	OrderID string `json:"order_id"`
	Message string `json:"message"`
}

func StartConsumers(ctx context.Context, p *Producer) {
	for _, topic := range GetTopics() {
		go startTopicConsumer(ctx, topic, config.Cfg.Kafka.GroupIDs.Notification, func(ctx context.Context, msg kafka.Message) {
			handleNotificaion(ctx, msg, p)
		})
	}
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

func handleNotificaion(ctx context.Context, msg kafka.Message, p *Producer) {
	orderID := string(msg.Key)

	var userID string
	var notificationMessage string

	var event NotifiableEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.Error("failed to unmarshal Kafka message body", "error", err)
		return
	}

	userID = event.UserID

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

	externalEvent := &NotificationEvent{
		UserID:  userID,
		OrderID: orderID,
		Message: notificationMessage,
	}

	eventBody, err := json.Marshal(externalEvent)
	if err != nil {
		slog.Error("failed to marshal notification for Kafka event", "error", err)
	} else {
		p.Produce(ctx, NotificationCreatedTopic, []byte(userID), eventBody)
	}

	slog.Info("notification formated and sended", "user_id", userID)
}
