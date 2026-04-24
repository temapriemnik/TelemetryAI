package transport

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"telemetryai/internal/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan interface{}
	register  chan *websocket.Conn
	unregister chan *websocket.Conn
	mu        sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan interface{}, 256),
		register:  make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				err := client.WriteJSON(message)
				if err != nil {
					client.Close()
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastLog(logEntry interface{}) {
	h.broadcast <- logEntry
}

func (h *Hub) BroadcastToProject(projectID int, logEntry interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		err := client.WriteJSON(map[string]interface{}{
			"project_id": projectID,
			"log":       logEntry,
		})
		if err != nil {
			client.Close()
		}
	}
}

type WSHandler struct {
	hub *Hub
}

func NewWSHandler(hub *Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

func (h *WSHandler) HandleWSSessions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["project_id"]

	token := r.URL.Query().Get("token")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int)
	_ = userID

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade error", "error", err)
		return
	}

	slog.Info("client connected to websocket", "project_id", projectID, "user_id", userID)

	h.hub.register <- conn

	go func() {
		defer func() {
			h.hub.unregister <- conn
			conn.Close()
		}()

		conn.SetReadLimit(512)
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetWriteDeadline(time.Now().Add(60 * time.Second))

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

type LogMessage struct {
	ID        int    `json:"id"`
	ProjectID int    `json:"project_id"`
	Level    string `json:"level"`
	Message string `json:"message"`
	Time    string `json:"time"`
}