package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	GRPC struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"grpc"`
	Kafka struct {
		Brokers  string `mapstructure:"brokers"`
		GroupIDs struct {
			Orders string `mapstructure:"orders"`
		} `mapstructure:"group_ids"`
		Topics struct {
			PaymentSucceeded string `mapstructure:"payment_succeeded"`
			PaymentFailed    string `mapstructure:"payment_failed"`
			PaymentRequested string `mapstructure:"payment_requested"`
		} `mapstructure:"topics"`
	} `mapstructure:"kafka"`
}

var Cfg Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./payments-service")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		slog.Warn("config file not found. Relying on environment variables.")
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		slog.Error("unable to decode config into struct", "error", err)
		os.Exit(1)
	}
}
