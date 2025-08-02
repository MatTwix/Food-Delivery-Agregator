package messaging

import (
	"context"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/config"
	"github.com/segmentio/kafka-go"
)

const RestaurantCreatedTopic = "restaurant.created"

type Producer struct {
	writer *kafka.Writer
}

func NewProducer() (*Producer, error) {
	cfg := config.LoadConfig()

	if cfg.KafkaBrokers == "" {
		log.Fatal("KAFKA_BROKERS environment variable is not set")
	}

	brokers := strings.Split(cfg.KafkaBrokers, ",")

	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		log.Printf("Error dialing Kafka broker: %v", err)
		return nil, err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Printf("Error getting Kafka controller: %v", err)
		return nil, err
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Printf("Error dialing Kafka controller: %v", err)
		return nil, err
	}
	defer controllerConn.Close()

	topicConfig := kafka.TopicConfig{
		Topic:             RestaurantCreatedTopic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}
	err = controllerConn.CreateTopics(topicConfig)
	if err != nil {
		log.Printf("Warning: topic cannot be created: %v", err)
	} else {
		log.Printf("Topic '%s' successfully created or already exists", RestaurantCreatedTopic)
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        RestaurantCreatedTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}

	return &Producer{writer: w}, nil
}

func (p *Producer) Produce(ctx context.Context, key, value []byte) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
	if err != nil {
		log.Printf("Failed to write message to Kafka: %v", err)
		return err
	}

	log.Printf("Message sent to topic %s", RestaurantCreatedTopic)

	return nil
}
