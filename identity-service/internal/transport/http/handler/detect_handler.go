package handler

import (
	"encoding/json"
	"net/http"

	"identity-service/internal/application/dto"
	"identity-service/internal/application/service"
)

type DetectHandler struct {
	apiKeyService *service.APIKeyService
}

func NewDetectHandler(apiKeyService *service.APIKeyService) *DetectHandler {
	return &DetectHandler{apiKeyService: apiKeyService}
}

func (h *DetectHandler) Detect(w http.ResponseWriter, r *http.Request) {
	var req dto.DetectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.APIKey == "" {
		http.Error(w, "api_key required", http.StatusBadRequest)
		return
	}

	key, err := h.apiKeyService.Verify(r.Context(), req.APIKey)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(dto.DetectResponse{
			Level:     "forbidden",
			ProjectID: "",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.DetectResponse{
		Level:     key.Level,
		ProjectID: key.ProjectID.String(),
	})
}