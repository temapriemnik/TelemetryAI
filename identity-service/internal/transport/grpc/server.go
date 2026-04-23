package grpc

import (
	"context"
	"encoding/json"

	"identity-service/internal/application/service"
	"identity-service/internal/domain/repository"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AddAPIKeyRequest struct {
	ApiKey    string `json:"api_key"`
	ProjectId string `json:"project_id"`
}

type AddAPIKeyResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type DeleteAPIKeyRequest struct {
	ApiKey string `json:"api_key"`
}

type DeleteAPIKeyResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ListAPIKeysRequest struct{}

type ListAPIKeysResponse struct {
	Keys []APIKeyPair `json:"keys"`
}

type APIKeyPair struct {
	ApiKey    string `json:"api_key"`
	ProjectId string `json:"project_id"`
}

type APIKeyServer struct {
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

	var pairs []APIKeyPair
	for _, k := range allKeys {
		pairs = append(pairs, APIKeyPair{
			ApiKey:    k.APIKey,
			ProjectId: k.ProjectID.String(),
		})
	}

	return &ListAPIKeysResponse{Keys: pairs}, nil
}

func (s *APIKeyServer) Register(grpcServer *grpc.Server) {
	grpcServer.RegisterService(&grpc.ServiceDesc{
		ServiceName: "grpc.APIKeyService",
		HandlerType: (*APIKeyServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "AddAPIKey",
				Handler:    _AddAPIKey_Handler,
			},
			{
				MethodName: "DeleteAPIKey",
				Handler:    _DeleteAPIKey_Handler,
			},
			{
				MethodName: "ListAPIKeys",
				Handler:    _ListAPIKeys_Handler,
			},
		},
		Streams: []grpc.StreamDesc{},
	}, s)
}

func _AddAPIKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	raw := json.RawMessage{}
	if err := dec(&raw); err != nil {
		return nil, err
	}
	var in AddAPIKeyRequest
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, err
	}
	return srv.(*APIKeyServer).AddAPIKey(ctx, &in)
}

func _DeleteAPIKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	raw := json.RawMessage{}
	if err := dec(&raw); err != nil {
		return nil, err
	}
	var in DeleteAPIKeyRequest
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, err
	}
	return srv.(*APIKeyServer).DeleteAPIKey(ctx, &in)
}

func _ListAPIKeys_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	raw := json.RawMessage{}
	if err := dec(&raw); err != nil {
		return nil, err
	}
	var in ListAPIKeysRequest
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, err
	}
	return srv.(*APIKeyServer).ListAPIKeys(ctx, &in)
}

type APIKeyServiceServer interface {
	AddAPIKey(ctx context.Context, req *AddAPIKeyRequest) (*AddAPIKeyResponse, error)
	DeleteAPIKey(ctx context.Context, req *DeleteAPIKeyRequest) (*DeleteAPIKeyResponse, error)
	ListAPIKeys(ctx context.Context, req *ListAPIKeysRequest) (*ListAPIKeysResponse, error)
}