package transport

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Router struct {
	authHandler     *AuthHandler
	projectHandler  *ProjectHandler
	logHandler      *LogHandler
	wsHandler      *WSHandler
}

func NewRouter(
	authHandler *AuthHandler,
	projectHandler *ProjectHandler,
	logHandler *LogHandler,
	wsHandler *WSHandler,
) *Router {
	return &Router{
		authHandler:    authHandler,
		projectHandler: projectHandler,
		logHandler:     logHandler,
		wsHandler:     wsHandler,
	}
}

func (r *Router) Setup(authMiddleware func(http.Handler) http.Handler) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/auth/register", r.authHandler.Register).Methods("POST")
	router.HandleFunc("/auth/login", r.authHandler.Login).Methods("POST")

	projectRouter := router.PathPrefix("/projects").Subrouter()
	projectRouter.Use(authMiddleware)
	projectRouter.HandleFunc("", r.projectHandler.Create).Methods("POST")
	projectRouter.HandleFunc("", r.projectHandler.GetAll).Methods("GET")
	projectRouter.HandleFunc("/{id}", r.projectHandler.GetByID).Methods("GET")
	projectRouter.HandleFunc("/{id}", r.projectHandler.Delete).Methods("DELETE")
	projectRouter.HandleFunc("/{id}", r.projectHandler.Update).Methods("PUT")

	logRouter := router.PathPrefix("").Subrouter()
	logRouter.HandleFunc("/logs", r.logHandler.Receive).Methods("POST")

	projectLogsRouter := router.PathPrefix("").Subrouter()
	projectLogsRouter.Use(authMiddleware)
	projectLogsRouter.HandleFunc("/projects/{project_id}/logs", r.logHandler.GetByProject).Methods("GET")

	wsRouter := router.PathPrefix("/ws").Subrouter()
	wsRouter.Use(authMiddleware)
	wsRouter.HandleFunc("/projects/{project_id}", r.wsHandler.HandleWSSessions).Methods("GET")

	return router
}