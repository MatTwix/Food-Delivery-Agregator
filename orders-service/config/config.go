package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	HTTP struct {
		Port string `mapstucture:"port"`
	} `mapstructure:"http"`
	GRPC struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"grpc"`
	DB struct {
		Source string `mapstructure:"source"`
	} `mapstructure:"db"`
	Kafka struct {
		Brokers  string `mapstructure:"brokers"`
		GroupIDs struct {
			Restaurants string `mapstructure:"restaurants"`
			Payments    string `mapstructure:"payments"`
			Couriers    string `mapstructure:"couriers"`
		} `mapstructure:"group_ids"`
		Topics struct {
			RestaurantCreated string `mapstructure:"restaurant_created"`
			RestaurantUpdated string `mapstructure:"restaurant_updated"`
			RestaurantDeleted string `mapstructure:"restaurant_deleted"`

			OrderCreated   string `mapstructure:"order_created"`
			OrderPaid      string `mapstructure:"order_paid"`
			OrderPickedUp  string `mapstructure:"order_picked_up"`
			OrderDelivered string `mapstructure:"order_delivered"`

			PaymentSucceeded string `mapstructure:"payment_succeeded"`
			PaymentFailed    string `mapstructure:"payment_failed"`
			PaymentRequested string `mapstructure:"payment_requested"`

			CourierRequested    string `mapstructure:"courier_requested"`
			CourierAssigned     string `mapstructure:"courier_assigned"`
			CourierSearchFailed string `mapstructure:"courier_search_failed"`
		} `mapstructure:"topics"`
	} `mapstructure:"kafka"`
}

var Cfg Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./orders-service")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		slog.Warn("config file not found. Relying on environment variables")
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		slog.Error("unable to decode config into struct", "error", err)
		os.Exit(1)
	}
}
