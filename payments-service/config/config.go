package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Kafka struct {
		Brokers string `mapstructure:"brokers"`
		Topics  struct {
			PaymentSucceeded string `mapstructure:"payment_succeeded"`
			PaymentFailed    string `mapstructure:"payment_failed"`
			OrderCreated     string `mapstructure:"order_created"`
			OrderUpdated     string `mapstructure:"order_updated"`
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
		log.Printf("Warning: config file not found. Relying on environment variables.")
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}
}
