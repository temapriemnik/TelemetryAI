package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	App     AppConfig       `yaml:"app"`
	NATS    NATSConfig     `yaml:"nats"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type AppConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	JWTSecret      string `yaml:"jwt_secret"`
	NLPServiceURL  string `yaml:"nlp_service_url"`
}

type NATSConfig struct {
	URLs        string `yaml:"urls"`
	Topic       string `yaml:"topic"`
	ErrorTopic string `yaml:"error_topic"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if v := os.Getenv("DATABASE_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DATABASE_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Database.Port)
	}
	if v := os.Getenv("DATABASE_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DATABASE_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DATABASE_DBNAME"); v != "" {
		cfg.Database.DBName = v
	}
	if v := os.Getenv("NLP_SERVICE_URL"); v != "" {
		cfg.App.NLPServiceURL = v
	}

	return &cfg, nil
}