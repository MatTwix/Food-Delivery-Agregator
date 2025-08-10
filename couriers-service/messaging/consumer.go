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

const (
	PaymentsGroupID = "orders-service-group-payments"

	CouriersGroupID = "couriers-service-group"
)

func StartConsumers(ctx context.Context, courierStore *store.CourierStore, p *Producer) {
	go startTopicConsumer(ctx, OrderPaidTopic, PaymentsGroupID, func(ctx context.Context, msg kafka.Message) {
		handleOrderPaid(ctx, msg, courierStore, p)
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

	var event OrderPaidEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("Error unmarshaling Kafka message: %v", err)
		return
	}
	log.Printf("Received message from topic %s: Key=%s, Value=%s", msg.Topic, string(msg.Key), string(msg.Value))

	courier, err := store.GetAvailable(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("No available couriers for order %s. Publishing failure event.", event.ID)
			p.Produce(ctx, CourierSearchFailedTopic, []byte(event.ID), msg.Value)
		} else {
			log.Printf("Error searching available courier: %v", err)
		}
		return
	}

	if err := store.UpdateStatus(ctx, courier.ID, "busy"); err != nil {
		log.Printf("Error updating courier status: %v", err)
		return
	}

	eventBody, err := json.Marshal(courier)
	if err != nil {
		log.Printf("Error marshaling courier for Kafka event: %v", err)
		return
	} else {
		p.Produce(ctx, CourierAssignedTopic, []byte(event.ID), eventBody)
	}

	log.Printf("Successfully assigned courier for order %s.", event.ID)
}
