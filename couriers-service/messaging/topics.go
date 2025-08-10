package messaging

import (
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/config"
	"github.com/segmentio/kafka-go"
)

type Topic string

const (
	OrderPaidTopic      Topic = "order.paid"
	OrderDeliveredTopic Topic = "order.delivered"

	CourierAssignedTopic     Topic = "courier.assigned"
	CourierSearchFailedTopic Topic = "courier.search.failed"
)

var Topics = []Topic{
	OrderPaidTopic,
	OrderDeliveredTopic,

	CourierAssignedTopic,
	CourierSearchFailedTopic,
}

func InitTopics() {
	cfg := config.LoadConfig()
	if cfg.KafkaBrokers == "" {
		log.Fatal("KAFKA_BROKERS environment variable is not set")
	}
	brokers := strings.Split(cfg.KafkaBrokers, ",")
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
			Topic:             string(topic),
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
