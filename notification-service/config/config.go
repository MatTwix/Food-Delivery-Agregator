package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Kafka struct {
		Brokers  string `mapstructure:"brokers"`
		GroupIDs struct {
			Notification string `mapstructure:"notification"`
		} `mapstructure:"group_ids"`
		Topics struct {
			PaymentSucceeded string `mapstructure:"payment_succeeded"`
			PaymentFailed    string `mapstructure:"payment_failed"`

			OrderCreated   string `mapstructure:"order_created"`
			OrderUpdated   string `mapstructure:"order_updated"`
			OrderPickedUp  string `mapstructure:"order_picked_up"`
			OrderDelivered string `mapstructure:"order_delivered"`
		} `mapstructure:"topics"`
	} `mapstructure:"kafka"`
}

var Cfg Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./notification-service")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		slog.Warn("config file not found. Relying on environment variables")
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		slog.Error("unable to devode config into struct", "error", err)
		os.Exit(1)
	}
}
