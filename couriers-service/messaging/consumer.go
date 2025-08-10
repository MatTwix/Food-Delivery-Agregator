package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/store"
	"github.com/jackc/pgx/v5"
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
	PaymentsGroupID = "couriers-service-group-payments"

	CouriersGroupID = "couriers-service-group"
)

func StartConsumers(ctx context.Context, courierStore *store.CourierStore, deliveryStore *store.DeliveryStore, p *Producer) {
	go startTopicConsumer(ctx, OrderPaidTopic, PaymentsGroupID, func(ctx context.Context, msg kafka.Message) {
		handleOrderPaid(ctx, msg, courierStore, p)
	})

	go startTopicConsumer(ctx, OrderDeliveredTopic, CouriersGroupID, func(ctx context.Context, msg kafka.Message) {
		handleOrderDelivered(ctx, msg, courierStore, deliveryStore)
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

func handleOrderPaid(ctx context.Context, msg kafka.Message, store *store.CourierStore, p *Producer) {
	log.Println("Handling 'order.paid' event...")

	var receivedEvent OrderPaidEvent
	if err := json.Unmarshal(msg.Value, &receivedEvent); err != nil {
		log.Printf("Error unmarshaling Kafka message: %v", err)
		return
	}
	log.Printf("Received message from topic %s: Key=%s, Value=%s", msg.Topic, string(msg.Key), string(msg.Value))

	courier, err := store.GetAvailable(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("No available couriers for order %s. Publishing failure event.", receivedEvent.ID)
			p.Produce(ctx, CourierSearchFailedTopic, []byte(receivedEvent.ID), msg.Value)
		} else {
			log.Printf("Error searching available courier: %v", err)
		}
		return
	}

	if err := store.UpdateStatus(ctx, courier.ID, "busy"); err != nil {
		log.Printf("Error updating courier status: %v", err)
		return
	}

	sendingEvent := CourierAssignedEvent{
		CourierID: receivedEvent.ID,
	}

	eventBody, err := json.Marshal(sendingEvent)
	if err != nil {
		log.Printf("Error marshaling courier for Kafka event: %v", err)
		return
	} else {
		p.Produce(ctx, CourierAssignedTopic, []byte(receivedEvent.ID), eventBody)
	}

	log.Printf("Successfully assigned courier for order %s.", receivedEvent.ID)
}

func handleOrderDelivered(ctx context.Context, msg kafka.Message, courierStore *store.CourierStore, deliveryStore *store.DeliveryStore) {
	log.Println("Handling 'order.delivered event...")

	var event OrderDeliveredEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("Error unmarshaling Kafka message: %v", err)
		return
	}

	courierID, err := deliveryStore.GetCourierIDByOrderID(ctx, event.OrderID)
	if err != nil {
		log.Printf("Error getting courier id: %v", err)
		return
	}

	//TODO: check if there is another deliveries by current courier

	if err := courierStore.UpdateStatus(ctx, courierID, "available"); err != nil {
		log.Printf("Error updating courier status to 'available': %v", err)
		return
	}

	//TODO: publish event to courier.became_available topic (optionaly)

	log.Printf("Courier %s became available.", courierID)
}
