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

	log.Printf("Connected to NATS at %s, topic: %s", cfg.NATS.URLs, cfg.NATS.Topic)

	log.Println("Discovering containers...")

	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("failed to get containers: %v", err)
		return
	}

	exclude := map[string]bool{
		"log-collector":  true,
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
		go followLogs(context.Background(), nc, name, cfg.NATS.Topic, cfg.APIKey)
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

func followLogs(ctx context.Context, nc *nats.Conn, containerName, topic, apiKey string) {
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

		if nc == nil || nc.IsClosed() {
			log.Printf("nats connection closed, cannot publish for %s", containerName)
			continue
		}

		if err := nc.Publish(topic, b); err != nil {
			log.Printf("publish error: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("scanner error for %s: %v", containerName, err)
	}

	_ = cmd.Wait()
}

func isStreamExistsErr(err error) bool {
	// not used with plain NATS but kept for compatibility if referenced elsewhere
	return false
}

type Config struct {
	APIKey string    `json:"api_key"`
	NATS   NATSConfig `json:"nats"`
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
	// keep commas as-is; original code replaced with same value
	cfg.NATS.URLs = strings.ReplaceAll(cfg.NATS.URLs, ",", ",")

	return &cfg, nil
}
