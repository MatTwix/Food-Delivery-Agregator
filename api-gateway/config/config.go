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
	URLs struct {
		RestaurantsService string `mapstructure:"restaurants_service"`
		OrdersService      string `mapstructure:"orders_service"`
		CouriersService    string `mapstructure:"couriers_service"`
	} `mapstructure:"urls"`
}

var Cfg Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./api-gateway")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Warning: config file not found. Relying on environment variables.")
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("Unable to devode config into struct %v", err)
	}
}
