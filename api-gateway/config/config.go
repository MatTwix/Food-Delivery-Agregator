package config

import "os"

type Config struct {
	Port string

	RestaurantsServiceUrl string
	OrdersServiceUrl      string
}

func LoadConfig() Config {
	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	return Config{
		Port: port,

		RestaurantsServiceUrl: os.Getenv("RESTAURANTS_SERVICE_URL"),
		OrdersServiceUrl:      os.Getenv("ORDERS_SERVICE_URL"),
	}
}
