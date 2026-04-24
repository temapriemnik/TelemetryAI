package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	log.SetOutput(os.Stdout)
	log.Println("Starting log-collector agent...")

	cfg, err := loadConfig("config.json")
	if err != nil {
		log.Printf("Failed to load config: %v, using defaults", err)
		cfg = &Config{
			NATS: NATSConfig{
				URLs:  "localhost:4222",
				Topic: "raw.logs",
			},
		}
	}

	nc, err := nats.Connect(cfg.NATS.URLs)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("Failed to create jetstream: %v", err)
	}

	ctx := context.Background()
	_, err = js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     "logs",
		Subjects: []string{cfg.NATS.Topic},
	})
	if err != nil && !isStreamExistsErr(err) {
		log.Fatalf("Failed to create stream: %v", err)
	}

	log.Printf("Connected to NATS at %s, topic: %s", cfg.NATS.URLs, cfg.NATS.Topic)

	log.Println("Discovering containers...")

	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("failed to get containers: %v", err)
		return
	}

	exclude := map[string]bool{
		"log-collector": true,
		"/log-collector": true,
	}

	started := 0
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		name := scanner.Text()
		if exclude[name] {
			continue
		}
		log.Printf("Starting to follow logs for: %s", name)
		go followLogs(ctx, js, name, cfg.NATS.Topic, cfg.APIKey)
		started++
	}

	if started == 0 {
		log.Println("No other containers found")
	}

	log.Println("Agent ready")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	<-quit
}

func followLogs(ctx context.Context, js jetstream.JetStream, containerName, topic, apiKey string) {
	cmd := exec.Command("docker", "logs", "-f", "--tail", "10", containerName)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("pipe error for %s: %v", containerName, err)
		return
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		log.Printf("start error for %s: %v", containerName, err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		entry := LogEntry{
			Container: containerName,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Message:   scanner.Text(),
			APIKey:    apiKey,
		}

		b, err := json.Marshal(entry)
		if err != nil {
			log.Printf("marshal error: %v", err)
			continue
		}

		_, err = js.Publish(ctx, topic, b)
		if err != nil {
			log.Printf("publish error: %v", err)
		}
	}

	cmd.Wait()
}

func isStreamExistsErr(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "stream name already exists"
}

type Config struct {
	APIKey string    `json:"api_key"`
	NATS  NATSConfig `json:"nats"`
}

type NATSConfig struct {
	URLs  string `json:"urls"`
	Topic string `json:"topic"`
}

type LogEntry struct {
	Container string `json:"container"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	APIKey    string `json:"api_key"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.NATS.Topic == "" {
		cfg.NATS.Topic = "raw.logs"
	}

	if urls := os.Getenv("NATS_URLS"); urls != "" {
		cfg.NATS.URLs = urls
	}
	if cfg.NATS.URLs == "" {
		cfg.NATS.URLs = "localhost:4222"
	}
	cfg.NATS.URLs = strings.ReplaceAll(cfg.NATS.URLs, ",", ",")

	return &cfg, nil
}