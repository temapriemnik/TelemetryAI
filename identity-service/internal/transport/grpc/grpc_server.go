package grpc

import (
	"context"

	"identity-service/internal/application/service"
	"identity-service/internal/domain/repository"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type APIKeyServer struct {
	UnimplementedAPIKeyServiceServer
	apiKeyService *service.APIKeyService
	projectRepo  repository.ProjectRepository
}

func NewAPIKeyServer(apiKeyService *service.APIKeyService, projectRepo repository.ProjectRepository) *APIKeyServer {
	return &APIKeyServer{
		apiKeyService: apiKeyService,
		projectRepo: projectRepo,
	}
}

func (s *APIKeyServer) AddAPIKey(ctx context.Context, req *AddAPIKeyRequest) (*AddAPIKeyResponse, error) {
	if req.ApiKey == "" || req.ProjectId == "" {
		return &AddAPIKeyResponse{Success: false, Error: "api_key and project_id required"}, nil
	}

	projectUUID, err := uuid.Parse(req.ProjectId)
	if err != nil {
		return &AddAPIKeyResponse{Success: false, Error: "invalid project_id"}, nil
	}

	err = s.apiKeyService.Create(ctx, projectUUID, req.ApiKey)
	if err != nil {
		return &AddAPIKeyResponse{Success: false, Error: err.Error()}, nil
	}

	return &AddAPIKeyResponse{Success: true, Error: ""}, nil
}

func (s *APIKeyServer) DeleteAPIKey(ctx context.Context, req *DeleteAPIKeyRequest) (*DeleteAPIKeyResponse, error) {
	if req.ApiKey == "" {
		return &DeleteAPIKeyResponse{Success: false, Error: "api_key required"}, nil
	}

	err := s.apiKeyService.Delete(ctx, req.ApiKey)
	if err != nil {
		return &DeleteAPIKeyResponse{Success: false, Error: err.Error()}, nil
	}

	return &DeleteAPIKeyResponse{Success: true, Error: ""}, nil
}

func (s *APIKeyServer) ListAPIKeys(ctx context.Context, req *ListAPIKeysRequest) (*ListAPIKeysResponse, error) {
	allKeys, err := s.apiKeyService.ListAll(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var pairs []*APIKeyPair
	for _, k := range allKeys {
		pairs = append(pairs, &APIKeyPair{
			ApiKey:    k.APIKey,
			ProjectId: k.ProjectID.String(),
		})
	}

	return &ListAPIKeysResponse{Keys: pairs}, nil
}