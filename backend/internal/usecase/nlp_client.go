package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"telemetryai/internal/models"
)

const defaultNLPURL = "http://nlp-service:5000"

type NLPClient interface {
	PredictLevel(log string) (string, error)
}

type nlpClient struct {
	baseURL string
	client  *http.Client
}

type predictRequest struct {
	Log string `json:"log"`
}

type predictResponse struct {
	Label string `json:"label"`
}

func NewNLPClient(url string) NLPClient {
	if url == "" {
		url = defaultNLPURL
	}
	return &nlpClient{
		baseURL: url,
		client:  &http.Client{},
	}
}

func (c *nlpClient) PredictLevel(log string) (string, error) {
	body, err := json.Marshal(predictRequest{Log: log})
	if err != nil {
		return "", err
	}

	resp, err := c.client.Post(c.baseURL+"/predict", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nlp service returned status %d", resp.StatusCode)
	}

	var result predictResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Label, nil
}

func mapNLPToLogLevel(label string) models.LogLevel {
	switch strings.ToLower(label) {
	case "error":
		return models.LogLevelError
	case "warn", "warning":
		return models.LogLevelWarn
	default:
		return models.LogLevelInfo
	}
}