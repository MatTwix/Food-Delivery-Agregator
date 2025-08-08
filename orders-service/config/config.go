package config

import "os"

type Config struct {
	Port     string
	GrpcPort string

	DbSource string

	KafkaBrokers string
}

func LoadConfig() Config {
	port := os.Getenv("PORT")

	if port == "" {
		port = "3002"
	}

	return Config{
		Port:     port,
		GrpcPort: os.Getenv("GRPC_SERVER"),

		DbSource: os.Getenv("DB_SOURCE"),

		KafkaBrokers: os.Getenv("KAFKA_BROKERS"),
	}
}
