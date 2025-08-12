package messaging

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/models"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/store"
	"github.com/segmentio/kafka-go"
)

type OrderPaidEvent struct {
	ID string `json:"id"`
}

type OrderDeliveredEvent struct {
	OrderID string `json:"order_id"`
}

type CourierAssignedEvent struct {
	CourierID string `json:"courier_id"`
}

//TODO: refactor some consumers: make order delivery status changing be provided by single consumer

func StartConsumers(ctx context.Context, restaurantStore *store.RestaurantStore, orderStore *store.OrderStore, p *Producer) {
	go startTopicConsumer(ctx, RestaurantCreatedTopic, config.Cfg.Kafka.GroupIDs.Restaurants, func(ctx context.Context, msg kafka.Message) {
		handleRestaurantCreated(ctx, msg, restaurantStore)
	})

	go startTopicConsumer(ctx, RestaurantUpdatedTopic, config.Cfg.Kafka.GroupIDs.Restaurants, func(ctx context.Context, msg kafka.Message) {
		handleRestaurantUpdated(ctx, msg, restaurantStore)
	})

	go startTopicConsumer(ctx, RestaurantDeletedTopic, config.Cfg.Kafka.GroupIDs.Restaurants, func(ctx context.Context, msg kafka.Message) {
		handleRestaurantDeleted(ctx, msg, restaurantStore)
	})

	go startTopicConsumer(ctx, PaymentSucceededTopic, config.Cfg.Kafka.GroupIDs.Payments, func(ctx context.Context, msg kafka.Message) {
		handlePaymentSucceeded(ctx, msg, orderStore, p)
	})

	go startTopicConsumer(ctx, PaymentFailedTopic, config.Cfg.Kafka.GroupIDs.Payments, func(ctx context.Context, msg kafka.Message) {
		handlePaymentFailed(ctx, msg, orderStore)
	})

	go startTopicConsumer(ctx, CourierAssignedTopic, config.Cfg.Kafka.GroupIDs.Couriers, func(ctx context.Context, msg kafka.Message) {
		handleCourierAssigned(ctx, msg, orderStore)
	})

	go startTopicConsumer(ctx, CourierSearchFailedTopic, config.Cfg.Kafka.GroupIDs.Couriers, func(ctx context.Context, msg kafka.Message) {
		handleCourierSearchFailed(ctx, msg, orderStore)
	})

	go startTopicConsumer(ctx, OrderPickedUpTopic, config.Cfg.Kafka.GroupIDs.Couriers, func(ctx context.Context, msg kafka.Message) {
		handleOrderPickedUp(ctx, msg, orderStore)
	})

	go startTopicConsumer(ctx, OrderDeliveredTopic, config.Cfg.Kafka.GroupIDs.Couriers, func(ctx context.Context, msg kafka.Message) {
		handleOrderDelivered(ctx, msg, orderStore)
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

	slog.Info("Starting Kafka consumer", "topic", topic, "group_id", groupID)

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

func handleRestaurantCreated(ctx context.Context, msg kafka.Message, store *store.RestaurantStore) {
	slog.Info("handling event", "event", RestaurantCreatedTopic)

	var restaurant models.Restaurant
	if err := json.Unmarshal(msg.Value, &restaurant); err != nil {
		slog.Error("failed to unmarshl Kafka message", "error", err)
		return
	}
	slog.Info("received message", "topic", msg.Topic, "key", string(msg.Key), "value", string(msg.Value))

	if err := store.Upsert(ctx, &restaurant); err != nil {
		slog.Error("failed to upsert restaurant", "error", err)
		return
	}

	slog.Info("successfully saved restaurant to the local database", "restaurant", restaurant.Name)
}

func handleRestaurantUpdated(ctx context.Context, msg kafka.Message, store *store.RestaurantStore) {
	//this handler copies previous one for remaining obility to modify the exact handler (send notifications on creating e.t.c)

	slog.Info("handling event", "event", RestaurantUpdatedTopic)

	var restaurant models.Restaurant
	if err := json.Unmarshal(msg.Value, &restaurant); err != nil {
		slog.Error("failed to unmarshl Kafka message", "error", err)
		return
	}
	slog.Info("received message", "topic", msg.Topic, "key", string(msg.Key), "value", string(msg.Value))

	if err := store.Upsert(ctx, &restaurant); err != nil {
		slog.Error("failed to upsert restaurant", "error", err)
		return
	}

	slog.Info("successfully saved restaurant to the local database", "restaurant", restaurant.Name)
}

func handleRestaurantDeleted(ctx context.Context, msg kafka.Message, store *store.RestaurantStore) {
	slog.Info("handling event", "topic", RestaurantDeletedTopic)

	var restaurant models.Restaurant
	if err := json.Unmarshal(msg.Value, &restaurant); err != nil {
		slog.Error("failed to unmarshl Kafka message", "error", err)
		return
	}
	slog.Info("received message", "topic", msg.Topic, "key", string(msg.Key), "value", string(msg.Value))

	if err := store.Delete(ctx, restaurant.ID); err != nil {
		slog.Error("failed to delete restaurant", "error", err)
		return
	}

	slog.Info("successfully deleted restaurant from the local database", "restaurant_id", restaurant.ID)
}

func handlePaymentSucceeded(ctx context.Context, msg kafka.Message, store *store.OrderStore, p *Producer) {
	orderID := string(msg.Key)
	slog.Info("handeling event", "event", PaymentSucceededTopic, "order_id", orderID)

	if err := store.UpdateStatus(ctx, orderID, "paid"); err != nil {
		slog.Error("failed to update order status to 'paid'", "order_id", orderID, "error", err)
		return
	}

	event := OrderPaidEvent{
		ID: orderID,
	}

	eventBody, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal message for Kafka event", "error", err)
		return
	} else {
		p.Produce(ctx, OrderPaidTopic, []byte(orderID), eventBody)
	}

	slog.Info("order status updated to 'paid'", "order_id", orderID)
}

func handlePaymentFailed(ctx context.Context, msg kafka.Message, store *store.OrderStore) {
	orderID := string(msg.Key)
	slog.Info("handeling event", "topic", PaymentFailedTopic, "order_id", orderID)

	if err := store.UpdateStatus(ctx, orderID, "payment_failed"); err != nil {
		slog.Error("failed to update order status to 'payment_failed'", "order_id", orderID, "error", err)
		return
	}

	slog.Info("order status updated to 'payment_failed'", "order_id", orderID)

	// Make event for refund or notifier
}

func handleCourierAssigned(ctx context.Context, msg kafka.Message, store *store.OrderStore) {
	orderID := string(msg.Key)
	slog.Info("handling event", "event", CourierAssignedTopic, "order_id", orderID)

	var event CourierAssignedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.Error("failed to unmarshal Kafka message", "error", err)
		return
	}

	if err := store.AssignCourier(ctx, orderID, event.CourierID); err != nil {
		slog.Error("failed to assign courier", "error", err)
		return
	}

	slog.Info("courier successfully assigned")
}

func handleCourierSearchFailed(ctx context.Context, msg kafka.Message, store *store.OrderStore) {
	orderID := string(msg.Key)
	slog.Info("handling event", "topic", CourierSearchFailedTopic, "order_id", orderID)

	if err := store.UpdateStatus(ctx, orderID, "no_couriers_available"); err != nil {
		slog.Error("failed to update order status to 'no_available_couriers'", "order_id", orderID, "error", err)
		return
	}

	slog.Info("order status updated to 'no_couriers_available'", "order_id", orderID)
}

func handleOrderPickedUp(ctx context.Context, msg kafka.Message, store *store.OrderStore) {
	orderID := string(msg.Key)
	slog.Info("handling event", "topic", OrderPickedUpTopic, "order_id", orderID)

	if err := store.UpdateStatus(ctx, orderID, "picked_up"); err != nil {
		slog.Error("failed to update order status to 'picked_up'", "order_id", orderID, "error", err)
		return
	}

	slog.Info("order status updated to 'picked_up'", "order_id", orderID)
}

func handleOrderDelivered(ctx context.Context, msg kafka.Message, store *store.OrderStore) {
	orderID := string(msg.Key)
	slog.Info("handling event", "event", OrderDeliveredTopic, "order_id", orderID)

	if err := store.UpdateStatus(ctx, orderID, "delivered"); err != nil {
		slog.Error("failed to update order status to 'delivered'", "error", err)
		return
	}

	slog.Info("order status updated to 'delivered'", "order_id", orderID)
}
