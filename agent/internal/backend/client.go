package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AlertData struct {
	UserEmail    string `json:"user_email"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Client struct {
	baseURL string
	client *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetAlertData(projectID int) (*AlertData, error) {
	url := fmt.Sprintf("%s/internal/projects/%d/alert-data", c.baseURL, projectID)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data AlertData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}