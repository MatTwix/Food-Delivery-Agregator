package config

import "os"

type Config struct {
	Port string

	DbSource string
}

func LoadConfig() Config {
	port := os.Getenv("PORT")

	if port == "" {
		port = "3001"
	}

	return Config{
		Port: port, DbSource: os.Getenv("DB_SOURCE"),
	}
}
