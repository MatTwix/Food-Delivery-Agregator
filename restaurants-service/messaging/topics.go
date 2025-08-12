package messaging

import (
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/config"
	"github.com/segmentio/kafka-go"
)

var (
	RestaurantCreatedTopic string
	RestaurantUpdatedTopic string
	RestaurantDeletedTopic string
)

var Topics = []string{
	RestaurantCreatedTopic,
	RestaurantUpdatedTopic,
	RestaurantDeletedTopic,
}

func InitTopicsNames() {
	RestaurantCreatedTopic = config.Cfg.Kafka.Topics.RestaurantCreated
	RestaurantUpdatedTopic = config.Cfg.Kafka.Topics.RestaurantUpdated
	RestaurantDeletedTopic = config.Cfg.Kafka.Topics.RestaurantDeleted
}

func InitTopics() {
	if config.Cfg.Kafka.Brokers == "" {
		log.Fatal("KAFKA_BROKERS environment variable is not set")
	}
	brokers := strings.Split(config.Cfg.Kafka.Brokers, ",")
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		log.Fatalf("Error dialing Kafka broker: %v", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Fatalf("Error getting Kafka controller: %v", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Fatalf("Error dialing Kafka controller: %v", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{}
	for _, topic := range Topics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		log.Printf("Warning: topic cannot be created: %v", err)
	} else {
		log.Printf("Topics successfully created or already exist")
	}
}
