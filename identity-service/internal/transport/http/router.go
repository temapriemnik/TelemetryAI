package http

import (
	"log/slog"

	"identity-service/internal/application/service"
	"identity-service/internal/transport/http/handler"
	"identity-service/internal/transport/http/middleware"

	"github.com/gorilla/mux"
)

func New(authHandler *handler.AuthHandler, projectHandler *handler.ProjectHandler, detectHandler *handler.DetectHandler, authService *service.AuthService, log *slog.Logger) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	r.Use(middleware.Logging(log))
	r.Use(middleware.PanicRecovery(log))
	r.Use(middleware.CORS)

	r.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	protected := r.PathPrefix("/projects").Subrouter()
	protected.Use(middleware.Auth(authService, log))
	protected.HandleFunc("", projectHandler.Create).Methods("POST")
	protected.HandleFunc("", projectHandler.List).Methods("GET")
	protected.HandleFunc("/{id}", projectHandler.Get).Methods("GET")
	protected.HandleFunc("/{id}/apikey", projectHandler.GetAPIKey).Methods("GET")

	r.HandleFunc("/detect", detectHandler.Detect).Methods("POST")

	return r
}