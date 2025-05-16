package server

// Package server implements HTTP handlers for reading and updating the selected save game setting
// from the server's configuration file.

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/snarf-dev/fsm/v2/internal/helpers"
)

// handleGetSettings returns the currently selected save game name from the configuration as JSON.
func (s *RestServer) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"save": s.fsmConfig.Factorio.Save,
	})
}

// handleUpdateSave updates the selected save game name in the configuration based on a JSON payload.
// The new save value is written to the config file and logged.
func (s *RestServer) handleUpdateSave(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Save string `json:"save"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		helpers.RenderErrorJSON(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	s.fsmConfig.Factorio.Save = payload.Save

	err := s.fsmConfig.SaveToFile()
	if err != nil {
		log.Printf("failed to update %s, %v\n", s.fsmConfig.Path, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to save config")
		return
	}

	log.Printf("save game file changed to %s\n", payload.Save)
	w.WriteHeader(http.StatusNoContent)
}
