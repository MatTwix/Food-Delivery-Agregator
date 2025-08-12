package config

import (
	"log/slog"
	"os"
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
		slog.Warn("config file not found. Relying on environment variables")
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		slog.Error("unable to devode config into struct", "error", err)
		os.Exit(1)
	}
}
