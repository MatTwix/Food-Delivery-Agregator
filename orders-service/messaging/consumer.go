package messaging

import (
	"context"
	"log"
	"strings"

	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/config"
	"github.com/segmentio/kafka-go"
)

const (
	RestaurantCreatedTopic = "restaurant.created"
	GroupID                = "orders-service-group"
)

func StartConsumer(ctx context.Context) {
	cfg := config.LoadConfig()

	if cfg.KafkaBrokers == "" {
		log.Fatal("KAFKA_BROKERS environment variable is not set")
	}

	brokers := strings.Split(cfg.KafkaBrokers, ",")

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  GroupID,
		Topic:    RestaurantCreatedTopic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	log.Printf("Starting Kafka consumer for topic '%s'", RestaurantCreatedTopic)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping Kafka consumer...")
				r.Close()
				return
			default:
				m, err := r.ReadMessage(ctx)
				if err != nil {
					log.Printf("Error reading message from Kafka: %v", err)
					continue
				}
				log.Printf("Received message from topic %s: Key=%s, Value=%s", m.Topic, string(m.Key), string(m.Value))
			}
		}
	}()
}
