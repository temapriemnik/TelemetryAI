package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	HTTPPort   string
	GRPCPort   string
}

func Load() *Config {
	return &Config{
		DBHost:     getenv("DB_HOST", "localhost"),
		DBPort:     getenvInt("DB_PORT", 5432),
		DBUser:     getenv("DB_USER", "identity"),
		DBPassword: getenv("DB_PASSWORD", "identity"),
		DBName:     getenv("DB_NAME", "identity"),
		HTTPPort:   getenv("HTTP_PORT", "8080"),
		GRPCPort:   getenv("GRPC_PORT", "50051"),
	}
}

func getenv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getenvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		var intVal int
		if _, err := fmt.Sscanf(v, "%d", &intVal); err == nil {
			return intVal
		}
	}
	return defaultValue
}