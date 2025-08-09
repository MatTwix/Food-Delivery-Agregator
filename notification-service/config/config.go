package config

import "os"

type Config struct {
	KafkaBrokers string
}

func LoadConfig() Config {
	return Config{
		KafkaBrokers: os.Getenv("KAFKA_BROKERS"),
	}
}
