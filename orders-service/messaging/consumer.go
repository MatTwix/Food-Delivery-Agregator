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

const (
	GroupID = "orders-service-group"
)

func StartConsumers(ctx context.Context, restaurantStore *store.RestaurantStore) {
	go startTopicConsumer(ctx, RestaurantCreatedTopic, GroupID, func(ctx context.Context, msg kafka.Message) {
		handleRestaurantCreated(ctx, msg, restaurantStore)
	})

	go startTopicConsumer(ctx, RestaurantUpdatedTopic, GroupID, func(ctx context.Context, msg kafka.Message) {
		handleRestaurantUpdated(ctx, msg, restaurantStore)
	})

	go startTopicConsumer(ctx, RestaurantDeletedTopic, GroupID, func(ctx context.Context, msg kafka.Message) {
		handleRestaurantDeleted(ctx, msg, restaurantStore)
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
	log.Printf("Handling 'restaurant.created' event...")

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
