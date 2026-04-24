package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"telemetryai/internal/middleware"
	"telemetryai/internal/usecase"
)

type LogHandler struct {
	logService *usecase.LogService
	hub       *Hub
}

func NewLogHandler(logService *usecase.LogService, hub *Hub) *LogHandler {
	return &LogHandler{logService: logService, hub: hub}
}

type ReceiveLogRequest struct {
	APIKey    string    `json:"api_key"`
	Timestamp time.Time `json:"timestamp"`
	Message  string    `json:"message"`
}

func (h *LogHandler) Receive(w http.ResponseWriter, r *http.Request) {
	var input ReceiveLogRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	if input.Timestamp.IsZero() {
		input.Timestamp = time.Now()
	}

	logInput := usecase.ReceiveLogInput{
		APIKey:   input.APIKey,
		Timestamp: input.Timestamp,
		Message: input.Message,
	}

	output, err := h.logService.Receive(logInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(output)

	if h.hub != nil {
		h.hub.BroadcastToProject(output.ProjectID, map[string]interface{}{
			"id":         output.ID,
			"project_id": output.ProjectID,
			"level":      output.Level,
			"message":   output.Message,
			"time":       output.Timestamp.Format(time.RFC3339),
		})
	}
}

func (h *LogHandler) GetByProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := strconv.Atoi(vars["project_id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid project id"})
		return
	}

	userIDVal := r.Context().Value(middleware.UserIDKey)
	if userIDVal == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
		return
	}
	userID := userIDVal.(int)

	logs, err := h.logService.GetByProjectID(projectID, userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(logs)
}