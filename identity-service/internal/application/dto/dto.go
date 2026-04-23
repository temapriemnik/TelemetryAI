package dto

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Token string `json:"token"`
	User  UserDTO `json:"user"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  UserDTO `json:"user"`
}

type UserDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type CreateProjectRequest struct {
	Name string `json:"name"`
}

type ProjectDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DetectRequest struct {
	APIKey string `json:"api_key"`
}

type DetectResponse struct {
	Level     string `json:"level"`
	ProjectID string `json:"project_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}