package apikeys

import (
	"context"
	"os"
	"testing"
)

func TestBadgerStorage(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewBadgerStorage(tmpDir)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	t.Run("Add and Get", func(t *testing.T) {
		err := storage.Add(ctx, "key1", "proj-uuid-1")
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		projectID, err := storage.Get(ctx, "key1")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if projectID != "proj-uuid-1" {
			t.Errorf("expected proj-uuid-1, got %s", projectID)
		}
	})

	t.Run("Get non-existent", func(t *testing.T) {
		_, err := storage.Get(ctx, "nonexistent")
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		storage.Add(ctx, "key2", "proj-uuid-2")
		err := storage.Delete(ctx, "key2")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		_, err = storage.Get(ctx, "key2")
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		storage.Add(ctx, "key3", "proj-uuid-3")
		storage.Add(ctx, "key4", "proj-uuid-4")

		keys, err := storage.List(ctx)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(keys) < 3 {
			t.Errorf("expected at least 3 keys, got %d", len(keys))
		}
	})

	t.Run("Persist after restart", func(t *testing.T) {
		storage.Close()

		storage2, err := NewBadgerStorage(tmpDir)
		if err != nil {
			t.Fatalf("failed to reopen storage: %v", err)
		}
		defer storage2.Close()

		projectID, err := storage2.Get(ctx, "key1")
		if err != nil {
			t.Fatalf("Get after reopen failed: %v", err)
		}
		if projectID != "proj-uuid-1" {
			t.Errorf("expected proj-uuid-1, got %s", projectID)
		}
	})
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}