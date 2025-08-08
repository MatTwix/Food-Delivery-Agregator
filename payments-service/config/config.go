package config

import "os"

type Config struct {
	Port string

	KafkaBrokers string
}

func LoadConfig() Config {
	port := os.Getenv("PORT")

	if port == "" {
		port = "3003"
	}

	return Config{
		Port:         port,
		KafkaBrokers: os.Getenv("KAFKA_BROKERS"),
	}
}
