package messaging

import (
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/MatTwix/Food-Delivery-Agregator/payments-service/config"
	"github.com/segmentio/kafka-go"
)

var (
	PaymentSucceededTopic string
	PaymentFailedTopic    string

	OrderCreatedTopic string
	OrderUpdatedTopic string
)

var Topics = []string{
	PaymentSucceededTopic,
	PaymentFailedTopic,

	OrderCreatedTopic,
	OrderUpdatedTopic,
}

func InitTopicsNames() {
	PaymentSucceededTopic = config.Cfg.Kafka.Topics.PaymentSucceeded
	PaymentFailedTopic = config.Cfg.Kafka.Topics.PaymentFailed
	OrderCreatedTopic = config.Cfg.Kafka.Topics.OrderCreated
	OrderUpdatedTopic = config.Cfg.Kafka.Topics.OrderUpdated
}

func InitTopics() {
	if config.Cfg.Kafka.Brokers == "" {
		slog.Error("KAFKA_BROKERS environment variable is not set")
		os.Exit(1)
	}
	brokers := strings.Split(config.Cfg.Kafka.Brokers, ",")
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		slog.Error("failed to dial Kafka broker", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		slog.Error("failed to get Kafka controller", "error", err)
		os.Exit(1)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		slog.Error("Error dialing Kafka controller", "error", err)
		os.Exit(1)
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
		slog.Warn("topic cannot be created", "error", err)
	} else {
		slog.Info("topics successfully created or already exist")
	}
}
