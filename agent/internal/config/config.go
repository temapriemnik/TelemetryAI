package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	NATS  NATSConfig  `yaml:"nats"`
	Backend BackendConfig `yaml:"backend"`
	SMTP  SMTPConfig  `yaml:"smtp"`
}

type NATSConfig struct {
	URLs string `yaml:"urls"`
	Topic string `yaml:"topic"`
}

type BackendConfig struct {
	URL string `yaml:"url"`
}

type SMTPConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	From string `yaml:"from"`
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

	return &cfg, nil
}