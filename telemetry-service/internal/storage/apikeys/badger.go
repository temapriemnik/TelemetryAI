package apikeys

import (
	"context"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

var ErrNotFound = errors.New("api key not found")

type Storage interface {
	Add(ctx context.Context, apiKey, projectID string) error
	Get(ctx context.Context, apiKey string) (string, error)
	Delete(ctx context.Context, apiKey string) error
	List(ctx context.Context) (map[string]string, error)
	Close() error
}

type BadgerStorage struct {
	db *badger.DB
}

func NewBadgerStorage(path string) (*BadgerStorage, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger: %w", err)
	}
	return &BadgerStorage{db: db}, nil
}

func (s *BadgerStorage) Add(ctx context.Context, apiKey, projectID string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(apiKey), []byte(projectID))
	})
}

func (s *BadgerStorage) Get(ctx context.Context, apiKey string) (string, error) {
	var projectID string
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(apiKey))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrNotFound
			}
			return err
		}
		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		projectID = string(valCopy)
		return nil
	})
	if err != nil {
		return "", err
	}
	return projectID, nil
}

func (s *BadgerStorage) Delete(ctx context.Context, apiKey string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(apiKey))
	})
}

func (s *BadgerStorage) List(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string)
	err := s.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()
			valCopy, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			result[string(item.Key())] = string(valCopy)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *BadgerStorage) Close() error {
	return s.db.Close()
}