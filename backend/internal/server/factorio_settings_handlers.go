package server

// Package server implements HTTP handlers for interacting with Factorio server settings,
// including reading and updating the server-settings.json configuration file.

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/snarf-dev/fsm/v2/internal/helpers"
)

// handleGetServerSettings responds with the contents of the Factorio server-settings.json file.
// It returns the raw JSON as-is from the configured file path.
func (s *RestServer) handleGetServerSettings(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean(s.fsmConfig.Factorio.Files.ServerSettings)
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Failed to read %s: %v\n", path, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to read server settings")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// handleUpdateServerSettings accepts a JSON payload containing updates to the server settings.
// It merges the new values with the existing file while preserving unmodified comments and ordering.
func (s *RestServer) handleUpdateServerSettings(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	path := filepath.Clean(s.fsmConfig.Factorio.Files.ServerSettings)
	originalData, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Failed to read %s: %v\n", path, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to read original server settings")
		return
	}

	var original map[string]json.RawMessage
	if err := json.Unmarshal(originalData, &original); err != nil {
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to parse original settings")
		return
	}

	var updated map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		helpers.RenderErrorJSON(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	for k, v := range updated {
		jsonVal, err := json.Marshal(v)
		if err != nil {
			helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to encode updated field: "+k)
			return
		}
		original[k] = jsonVal
	}

	sorted := make(map[string]json.RawMessage)
	keys := make([]string, 0, len(original))
	for k := range original {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sorted[k] = original[k]
	}
	finalData, err := json.MarshalIndent(sorted, "", "  ")
	if err != nil {
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to re-encode settings")
		return
	}

	if err := os.WriteFile(path, finalData, 0644); err != nil {
		log.Printf("Failed to write to %s: %v\n", path, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to write server settings")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGetFactorioUserSettings returns the configured Factorio.com username and token
// as a JSON object from the loaded FSM configuration.
func (s *RestServer) handleGetFactorioUserSettings(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"username": s.manager.cfg.Factorio.Username,
		"token":    s.manager.cfg.Factorio.Token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUpdateFactorioUserSettings accepts a JSON payload with a Factorio.com username and token,
// updates the current configuration, and persists the changes to disk.
func (s *RestServer) handleUpdateFactorioUserSettings(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username"`
		Token    string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		helpers.RenderErrorJSON(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	s.fsmConfig.Factorio.Username = payload.Username
	s.fsmConfig.Factorio.Token = payload.Token

	err := s.fsmConfig.SaveToFile()
	if err != nil {
		log.Printf("failed to update %s, %v\n", s.fsmConfig.Path, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to save config")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
