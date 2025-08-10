package messaging

import (
	"context"
	"encoding/json"
	"log"
	"strings"

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

const (
	RestaurantsGroupID = "orders-service-group-restaurants"
	PaymentsGroupID    = "orders-service-group-payments"
	CouriersGroupID    = "orders-service-group-couriers"
)

//TODO: refactor some consumers: make order delivery status changing be provided by single consumer

func StartConsumers(ctx context.Context, restaurantStore *store.RestaurantStore, orderStore *store.OrderStore, p *Producer) {
	go startTopicConsumer(ctx, RestaurantCreatedTopic, RestaurantsGroupID, func(ctx context.Context, msg kafka.Message) {
		handleRestaurantCreated(ctx, msg, restaurantStore)
	})

	go startTopicConsumer(ctx, RestaurantUpdatedTopic, RestaurantsGroupID, func(ctx context.Context, msg kafka.Message) {
		handleRestaurantUpdated(ctx, msg, restaurantStore)
	})

	go startTopicConsumer(ctx, RestaurantDeletedTopic, RestaurantsGroupID, func(ctx context.Context, msg kafka.Message) {
		handleRestaurantDeleted(ctx, msg, restaurantStore)
	})

	go startTopicConsumer(ctx, PaymentSucceededTopic, PaymentsGroupID, func(ctx context.Context, msg kafka.Message) {
		handlePaymentSucceeded(ctx, msg, orderStore, p)
	})

	go startTopicConsumer(ctx, PaymentFailedTopic, PaymentsGroupID, func(ctx context.Context, msg kafka.Message) {
		handlePaymentFailed(ctx, msg, orderStore)
	})

	go startTopicConsumer(ctx, CourierAssignedTopic, CouriersGroupID, func(ctx context.Context, msg kafka.Message) {
		handleCourierAssigned(ctx, msg, orderStore)
	})

	go startTopicConsumer(ctx, CourierSearchFailedTopic, CouriersGroupID, func(ctx context.Context, msg kafka.Message) {
		handleCourierSearchFailed(ctx, msg, orderStore)
	})

	go startTopicConsumer(ctx, OrderDeliveredTopic, CouriersGroupID, func(ctx context.Context, msg kafka.Message) {
		handleOrderDelivered(ctx, msg, orderStore)
	})
}

func startTopicConsumer(ctx context.Context, topic Topic, groupID string, handler func(ctx context.Context, msg kafka.Message)) {
	cfg := config.LoadConfig()

	if cfg.KafkaBrokers == "" {
		log.Fatal("KAFKA_BROKERS environment variable is not set")
	}

	brokers := strings.Split(cfg.KafkaBrokers, ",")

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    string(topic),
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	log.Printf("Starting Kafka consumer for topic '%s'", topic)

	go func() {
		defer r.Close()

		for {
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					break
				}
				log.Printf("Error reading message from topic '%s': %v", topic, err)
				continue
			}
			handler(ctx, m)
		}
		log.Printf("Stopping consumer for topic '%s'", topic)
	}()
}

func handleRestaurantCreated(ctx context.Context, msg kafka.Message, store *store.RestaurantStore) {
	log.Printf("[Consumer ID: %p] Handling 'restaurant.created' event...", store)

	var restaurant models.Restaurant
	if err := json.Unmarshal(msg.Value, &restaurant); err != nil {
		log.Printf("Error unmarshling Kafka message: %v", err)
		return
	}
	log.Printf("Received message from topic %s: Key=%s, Value=%s", msg.Topic, string(msg.Key), string(msg.Value))

	if err := store.Upsert(ctx, &restaurant); err != nil {
		log.Printf("Error upserting restaurant: %v", err)
		return
	}

	log.Printf("Successfully saved restaurant '%s' to the local database.", restaurant.Name)
}

func handleRestaurantUpdated(ctx context.Context, msg kafka.Message, store *store.RestaurantStore) {
	//this handler copies previous one for remaining obility to modify the exact handler (send notifications on creating e.t.c)

	log.Printf("Handling 'restaurant.updated' event...")

	var restaurant models.Restaurant
	if err := json.Unmarshal(msg.Value, &restaurant); err != nil {
		log.Printf("Error unmarshling Kafka message: %v", err)
		return
	}
	log.Printf("Received message from topic %s: Key=%s, Value=%s", msg.Topic, string(msg.Key), string(msg.Value))

	if err := store.Upsert(ctx, &restaurant); err != nil {
		log.Printf("Error upserting restaurant: %v", err)
		return
	}

	log.Printf("Successfully saved restaurant '%s' to the local database.", restaurant.Name)
}

func handleRestaurantDeleted(ctx context.Context, msg kafka.Message, store *store.RestaurantStore) {
	log.Printf("Handling 'restaurant.created' event...")

	var restaurant models.Restaurant
	if err := json.Unmarshal(msg.Value, &restaurant); err != nil {
		log.Printf("Error unmarshling Kafka message: %v", err)
		return
	}
	log.Printf("Received message from topic %s: Key=%s, Value=%s", msg.Topic, string(msg.Key), string(msg.Value))

	if err := store.Delete(ctx, restaurant.ID); err != nil {
		log.Printf("Error deleting restaurant: %v", err)
		return
	}

	log.Printf("Successfully deleted restaurant '%s' from the local database.", restaurant.ID)
}

func handlePaymentSucceeded(ctx context.Context, msg kafka.Message, store *store.OrderStore, p *Producer) {
	orderID := string(msg.Key)
	log.Printf("Handeling 'payment.succeeded' event for order ID: %s", orderID)

	if err := store.UpdateStatus(ctx, orderID, "paid"); err != nil {
		log.Printf("Error updating order status to 'paid' for order %s: %v", orderID, err)
		return
	}

	event := OrderPaidEvent{
		ID: orderID,
	}

	eventBody, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling message for Kafka event: %v", err)
		return
	} else {
		p.Produce(ctx, OrderPaidTopic, []byte(orderID), eventBody)
	}

	log.Printf("Order %s status updated to 'paid'.", orderID)
}

func handlePaymentFailed(ctx context.Context, msg kafka.Message, store *store.OrderStore) {
	orderID := string(msg.Key)
	log.Printf("Handeling 'payment.failed' event for order ID: %s", orderID)

	if err := store.UpdateStatus(ctx, orderID, "payment_failed"); err != nil {
		log.Printf("Error updating order status to 'payment_failed' for order %s: %v", orderID, err)
		return
	}

	log.Printf("Order %s status updated to 'payment_failed'.", orderID)

	// Make event for refund or notifier
}

func handleCourierAssigned(ctx context.Context, msg kafka.Message, store *store.OrderStore) {
	orderID := string(msg.Key)
	log.Printf("Handling 'courier.assigned' event for order ID: %s", orderID)

	var event CourierAssignedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("Error unmarshaling Kafka message: %v", err)
		return
	}

	if err := store.AssignCourier(ctx, orderID, event.CourierID); err != nil {
		log.Printf("Error assigning courier: %v", err)
		return
	}

	log.Printf("Courier successfully assigned.")
}

func handleCourierSearchFailed(ctx context.Context, msg kafka.Message, store *store.OrderStore) {
	orderID := string(msg.Key)
	log.Printf("Handling 'courier.search.failed' event for order ID: %s", orderID)

	if err := store.UpdateStatus(ctx, orderID, "no_couriers_available"); err != nil {
		log.Printf("Error updating order status to 'no_available_couriers' for order %s: %v", orderID, err)
		return
	}

	log.Printf("Order %s status updated to 'no_couriers_available.'", orderID)
}

func handleOrderDelivered(ctx context.Context, msg kafka.Message, store *store.OrderStore) {
	orderID := string(msg.Key)
	log.Printf("Handling 'order.delivered' event for order id: %s", orderID)

	if err := store.UpdateStatus(ctx, orderID, "delivered"); err != nil {
		log.Printf("Error updating order status to 'delivered': %v", err)
		return
	}

	log.Printf("Order %s status updated to 'delivered'.", orderID)
}
