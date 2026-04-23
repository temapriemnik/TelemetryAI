package grpc

import (
	"context"
	"testing"

	"telemetry-service/internal/storage/apikeys"
)

type mockStorage struct {
	keys map[string]string
}

func (m *mockStorage) Add(ctx context.Context, apiKey, projectID string) error {
	m.keys[apiKey] = projectID
	return nil
}

func (m *mockStorage) Get(ctx context.Context, apiKey string) (string, error) {
	if val, ok := m.keys[apiKey]; ok {
		return val, nil
	}
	return "", apikeys.ErrNotFound
}

func (m *mockStorage) Delete(ctx context.Context, apiKey string) error {
	delete(m.keys, apiKey)
	return nil
}

func (m *mockStorage) List(ctx context.Context) (map[string]string, error) {
	return m.keys, nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestServer(t *testing.T) {
	storage := &mockStorage{keys: make(map[string]string)}
	server := NewServer(storage)

	t.Run("AddAPIKey", func(t *testing.T) {
		resp, err := server.AddAPIKey(context.Background(), &AddAPIKeyRequest{
			ApiKey:    "test-key",
			ProjectId: "proj-123",
		})
		if err != nil {
			t.Fatalf("AddAPIKey failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success=true, got false")
		}
	})

	t.Run("Get after Add", func(t *testing.T) {
		resp, err := server.ListAPIKeys(context.Background(), &ListAPIKeysRequest{})
		if err != nil {
			t.Fatalf("ListAPIKeys failed: %v", err)
		}
		if len(resp.Keys) != 1 {
			t.Errorf("expected 1 key, got %d", len(resp.Keys))
		}
	})

	t.Run("DeleteAPIKey", func(t *testing.T) {
		resp, err := server.DeleteAPIKey(context.Background(), &DeleteAPIKeyRequest{
			ApiKey: "test-key",
		})
		if err != nil {
			t.Fatalf("DeleteAPIKey failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success=true")
		}
	})

	t.Run("Delete non-existent returns success", func(t *testing.T) {
		resp, err := server.DeleteAPIKey(context.Background(), &DeleteAPIKeyRequest{
			ApiKey: "nonexistent",
		})
		if err != nil {
			t.Fatalf("DeleteAPIKey failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success=true (idempotent)")
		}
	})
}