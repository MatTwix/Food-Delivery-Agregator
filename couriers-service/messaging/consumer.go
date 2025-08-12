package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/models"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/store"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/kafka-go"
)

type OrderPaidEvent struct {
	ID string `json:"id"`
}

type OrderPickedUpEvent struct {
	OrderID string `json:"order_id"`
}

type OrderDeliveredEvent struct {
	OrderID string `json:"order_id"`
}

type CourierAssignedEvent struct {
	CourierID string `json:"courier_id"`
}

func StartConsumers(ctx context.Context, courierStore *store.CourierStore, deliveryStore *store.DeliveryStore, p *Producer) {
	go startTopicConsumer(ctx, OrderPaidTopic, config.Cfg.Kafka.GroupIDs.Payments, func(ctx context.Context, msg kafka.Message) {
		handleOrderPaid(ctx, msg, courierStore, deliveryStore, p)
	})

	go startTopicConsumer(ctx, OrderDeliveredTopic, config.Cfg.Kafka.GroupIDs.Orders, func(ctx context.Context, msg kafka.Message) {
		handleOrderDelivered(ctx, msg, courierStore, deliveryStore)
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
			slog.Info("stopping consumer for due to context cancellation", "topic", topic)
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

func handleOrderPaid(ctx context.Context, msg kafka.Message, courierStore *store.CourierStore, deliveryStore *store.DeliveryStore, p *Producer) {
	slog.Info("handling event", "event", OrderPaidTopic)

	var receivedEvent OrderPaidEvent
	if err := json.Unmarshal(msg.Value, &receivedEvent); err != nil {
		slog.Error("failed to unmarshal Kafka message", "error", err)
		return
	}
	slog.Info("received message", "topic", msg.Topic, "key", string(msg.Key), "value", string(msg.Value))

	courier, err := courierStore.GetAvailable(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Info("no available couriers. Publishing failure event.", "order_id", receivedEvent.ID)
			p.Produce(ctx, CourierSearchFailedTopic, []byte(receivedEvent.ID), msg.Value)
		} else {
			slog.Error("failed to search available courier", "error", err)
		}
		return
	}

	delivery := models.Delivery{
		CourierID: courier.ID,
		OrderID:   receivedEvent.ID,
	}

	if err := deliveryStore.Create(ctx, &delivery); err != nil {
		slog.Error("failed to create delivery", "error", err)
		return
	}

	if err := courierStore.UpdateStatus(ctx, courier.ID, "busy"); err != nil {
		slog.Error("failed to update courier status", "error", err)
		return
	}

	sendingEvent := CourierAssignedEvent{
		CourierID: receivedEvent.ID,
	}

	eventBody, err := json.Marshal(sendingEvent)
	if err != nil {
		slog.Error("failed to marshal courier for Kafka event", "error", err)
		return
	} else {
		p.Produce(ctx, CourierAssignedTopic, []byte(receivedEvent.ID), eventBody)
	}

	slog.Info("successfully assigned courier", "order_id", receivedEvent.ID)
}

func handleOrderDelivered(ctx context.Context, msg kafka.Message, courierStore *store.CourierStore, deliveryStore *store.DeliveryStore) {
	slog.Info("handling event", "event", OrderDeliveredTopic)

	var event OrderDeliveredEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.Error("failed to unmarshal Kafka message", "error", err)
		return
	}

	courierID, err := deliveryStore.GetCourierIDByOrderID(ctx, event.OrderID)
	if err != nil {
		slog.Error("failed to get courier id", "error", err)
		return
	}

	//TODO: check if there is another deliveries by current courier

	if err := courierStore.UpdateStatus(ctx, courierID, "available"); err != nil {
		slog.Error("failed to update courier status to 'available'", "error", err)
		return
	}

	//TODO: publish event to courier.became_available topic (optionaly)

	slog.Info("courier became available", "courier_id", courierID)
}
