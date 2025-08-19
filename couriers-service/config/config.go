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
	DB struct {
		Source string `mapstructure:"source"`
	} `mapstructure:"db"`
	Kafka struct {
		Brokers  string `mapstructure:"brokers"`
		GroupIDs struct {
			Payments string `mapstructure:"payments"`
			Orders   string `mapstructure:"orders"`
			Users    string `mapstructure:"users"`
		} `mapstructure:"group_ids"`
		Topics struct {
			OrderPaid      string `mapstructure:"order_paid"`
			OrderPickedUp  string `mapstructure:"order_picked_up"`
			OrderDelivered string `mapstructure:"order_delivered"`

			CourierAssigned     string `mapstructure:"courier_assigned"`
			CourierSearchFailed string `mapstructure:"courier_search_failed"`

			UsersRoleAssigned string `mapstructure:"users_role_assigned"`
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
		slog.Warn("config file not found. Relying on environment variables")
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		slog.Error("unable to devode config into struct", "error", err)
		os.Exit(1)
	}
}
