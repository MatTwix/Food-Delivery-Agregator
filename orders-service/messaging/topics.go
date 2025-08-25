package messaging

import (
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/config"
	"github.com/segmentio/kafka-go"
)

var (
	RestaurantCreatedTopic string
	RestaurantUpdatedTopic string
	RestaurantDeletedTopic string

	OrderCreatedTopic   string
	OrderPaidTopic      string
	OrderPickedUpTopic  string
	OrderDeliveredTopic string

	PaymentSucceededTopic string
	PaymentFailedTopic    string

	CourierRequestedTopic    string
	CourierAssignedTopic     string
	CourierSearchFailedTopic string
)

var Topics []string

func InitTopicsNames() {
	RestaurantCreatedTopic = config.Cfg.Kafka.Topics.RestaurantCreated
	RestaurantUpdatedTopic = config.Cfg.Kafka.Topics.RestaurantUpdated
	RestaurantDeletedTopic = config.Cfg.Kafka.Topics.RestaurantDeleted

	OrderCreatedTopic = config.Cfg.Kafka.Topics.OrderCreated
	OrderPaidTopic = config.Cfg.Kafka.Topics.OrderPaid
	OrderPickedUpTopic = config.Cfg.Kafka.Topics.OrderPickedUp
	OrderDeliveredTopic = config.Cfg.Kafka.Topics.OrderDelivered

	PaymentSucceededTopic = config.Cfg.Kafka.Topics.PaymentSucceeded
	PaymentFailedTopic = config.Cfg.Kafka.Topics.PaymentFailed

	CourierRequestedTopic = config.Cfg.Kafka.Topics.CourierRequested
	CourierAssignedTopic = config.Cfg.Kafka.Topics.CourierAssigned
	CourierSearchFailedTopic = config.Cfg.Kafka.Topics.CourierSearchFailed

	Topics = []string{
		RestaurantCreatedTopic,
		RestaurantUpdatedTopic,
		RestaurantDeletedTopic,

		OrderCreatedTopic,
		OrderPaidTopic,
		OrderPickedUpTopic,
		OrderDeliveredTopic,

		PaymentSucceededTopic,
		PaymentFailedTopic,

		CourierRequestedTopic,
		CourierAssignedTopic,
		CourierSearchFailedTopic,
	}
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
		slog.Error("failed to dial Kafka controller", "error", err)
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
