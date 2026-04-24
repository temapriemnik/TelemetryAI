package transport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"telemetryai/internal/usecase"
)

type InternalHandler struct {
	projectService *usecase.ProjectService
}

func NewInternalHandler(projectService *usecase.ProjectService) *InternalHandler {
	return &InternalHandler{projectService: projectService}
}

func (h *InternalHandler) GetAlertData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid project id"})
		return
	}

	output, err := h.projectService.GetAlertData(projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(output)
}