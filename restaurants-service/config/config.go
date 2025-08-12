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
	GRPC struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"grpc"`
	DB struct {
		Source string `mapstructure:"source"`
	} `mapstructure:"db"`
	Kafka struct {
		Brokers string `mapstructure:"brokers"`
		Topics  struct {
			RestaurantCreated string `mapstructure:"restaurant_created"`
			RestaurantUpdated string `mapstructure:"restaurant_updated"`
			RestaurantDeleted string `mapstructure:"restaurant_deleted"`
		} `mapstructure:"topics"`
	} `mapstructure:"kafka"`
}

var Cfg Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./restaurants-service")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: config file not found. Relying on environment variables.")
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}
}
