package grpc

import (
	"context"
	"errors"

	apikeys "telemetry-service/internal/storage/apikeys"
)

type Server struct {
	UnimplementedAPIKeyServiceServer
	storage apikeys.Storage
}

func NewServer(storage apikeys.Storage) *Server {
	return &Server{storage: storage}
}

func (s *Server) AddAPIKey(ctx context.Context, req *AddAPIKeyRequest) (*AddAPIKeyResponse, error) {
	if req.ApiKey == "" || req.ProjectId == "" {
		return &AddAPIKeyResponse{Success: false, Error: "api_key and project_id are required"}, nil
	}
	if err := s.storage.Add(ctx, req.ApiKey, req.ProjectId); err != nil {
		return &AddAPIKeyResponse{Success: false, Error: err.Error()}, nil
	}
	return &AddAPIKeyResponse{Success: true}, nil
}

func (s *Server) DeleteAPIKey(ctx context.Context, req *DeleteAPIKeyRequest) (*DeleteAPIKeyResponse, error) {
	if req.ApiKey == "" {
		return &DeleteAPIKeyResponse{Success: false, Error: "api_key is required"}, nil
	}
	if err := s.storage.Delete(ctx, req.ApiKey); err != nil {
		if errors.Is(err, apikeys.ErrNotFound) {
			return &DeleteAPIKeyResponse{Success: false, Error: "api_key not found"}, nil
		}
		return &DeleteAPIKeyResponse{Success: false, Error: err.Error()}, nil
	}
	return &DeleteAPIKeyResponse{Success: true}, nil
}

func (s *Server) ListAPIKeys(ctx context.Context, req *ListAPIKeysRequest) (*ListAPIKeysResponse, error) {
	keys, err := s.storage.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*APIKeyPair, 0, len(keys))
	for k, v := range keys {
		result = append(result, &APIKeyPair{ApiKey: k, ProjectId: v})
	}
	return &ListAPIKeysResponse{Keys: result}, nil
}