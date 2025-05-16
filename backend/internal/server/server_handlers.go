// Package server provides HTTP endpoints for controlling and monitoring the Factorio server,
// including log streaming, start/stop controls, and status reporting.
package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/snarf-dev/fsm/v2/internal/helpers"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// return r.Header.Get("Origin") == "http://your-frontend-domain"
		return true
	},
}

// handleLogStream upgrades the HTTP connection to a WebSocket and streams log output
// from the running Factorio server to the connected client in real-time.
func (s *RestServer) handleLogStream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer conn.Close()

	logCh := s.manager.SubscribeToLogs()
	for msg := range logCh {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			break
		}
	}
}

// startHandler starts the Factorio server and responds with the updated status as JSON.
func (s *RestServer) startHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.manager.Start(); err != nil {
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to start the server")
		return
	}
	s.renderStatusJSON(w)
}

// stopHandler stops the Factorio server and responds with the updated status as JSON.
func (s *RestServer) stopHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.manager.Stop(); err != nil {
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to stop the server")
		return
	}
	s.renderStatusJSON(w)
}

// statusHandler returns the current server status as JSON.
func (s *RestServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	s.renderStatusJSON(w)
}

// renderStatusJSON encodes and writes the current server status as a JSON response.
func (s *RestServer) renderStatusJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.manager.Status())
}
