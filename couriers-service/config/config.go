package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	HTTP struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"http"`
	DB struct {
		Source string `mapstructure:"source"`
	} `mapstructure:"db"`
	Kafka struct {
		Brokers  string `mapstructure:"brokers"`
		GroupIDs struct {
			Payments string `mapstructure:"payments"`
			Orders   string `mapstructure:"orders"`
		} `mapstructure:"group_ids"`
		Topics struct {
			OrderPaid      string `mapstructure:"order_paid"`
			OrderPickedUp  string `mapstructure:"order_picked_up"`
			OrderDelivered string `mapstructure:"order_delivered"`

			CourierAssigned     string `mapstructure:"coutier_assigned"`
			CourierSearchFailed string `mapstructure:"coutier_search_failed"`
		} `mapstructure:"topics"`
	} `mapstructure:"kafka"`
}

var Cfg Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./couriers-service")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Warning: config file not found. Relying on environment variables.")
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("Unable to devode config into struct %v", err)
	}
}
