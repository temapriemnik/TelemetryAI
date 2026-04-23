package handler

import (
	"encoding/json"
	"net/http"

	"identity-service/internal/application/dto"
	"identity-service/internal/application/service"
	"identity-service/internal/domain/entity"
	"identity-service/internal/transport/http/middleware"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	token, user, err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrUserExists {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.RegisterResponse{
		Token: token,
		User: dto.UserDTO{
			ID:    user.ID.String(),
			Email: user.Email,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	token, user, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	resp := dto.LoginResponse{
		Token: token,
		User: dto.UserDTO{
			ID:    user.ID.String(),
			Email: user.Email,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type ProjectHandler struct {
	projectService *service.ProjectService
}

func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r).(*entity.User)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	project, err := h.projectService.Create(r.Context(), user.ID, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.ProjectDTO{
		ID:   project.ID.String(),
		Name: project.Name,
	})
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r).(*entity.User)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	projectID, ok := vars["id"]
	if !ok {
		http.Error(w, "project id required", http.StatusBadRequest)
		return
	}

	project, err := h.projectService.GetByID(r.Context(), uuid.MustParse(projectID), user.ID)
	if err != nil {
		if err == service.ErrProjectNotFound || err == service.ErrProjectAccess {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.ProjectDTO{
		ID:   project.ID.String(),
		Name: project.Name,
	})
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r).(*entity.User)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	projects, err := h.projectService.GetByUserID(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var result []dto.ProjectDTO
	for _, p := range projects {
		result = append(result, dto.ProjectDTO{
			ID:   p.ID.String(),
			Name: p.Name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *ProjectHandler) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r).(*entity.User)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	projectID, ok := vars["id"]
	if !ok {
		http.Error(w, "project id required", http.StatusBadRequest)
		return
	}

	key, err := h.projectService.GetAPIKey(r.Context(), uuid.MustParse(projectID), user.ID)
	if err != nil {
		if err == service.ErrProjectNotFound || err == service.ErrProjectAccess {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"api_key": key})
}