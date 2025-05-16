package server

// Package server provides HTTP handlers for interacting with the Factorio mod list,
// allowing users to retrieve and toggle mod states through the REST API.

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/snarf-dev/fsm/v2/internal/helpers"
	"github.com/snarf-dev/fsm/v2/internal/mods"
)

// modsHandler returns the full contents of mod-list.json as a JSON response.
func (s *RestServer) modsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(fmt.Sprintf("%s/mod-list.json", s.fsmConfig.Factorio.ModsDir))
	if err != nil {
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to read mod list")
		return
	}
	var modList mods.ModList
	if err := json.Unmarshal(data, &modList); err != nil {
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to parse mod list")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(modList)
}

// toggleModHandler updates the enabled state of a specific mod based on query parameters.
// It returns the updated mod list as JSON.
func (s *RestServer) toggleModHandler(w http.ResponseWriter, r *http.Request) {
	modName := r.URL.Query().Get("mod")
	enabledStr := r.URL.Query().Get("enabled")
	enabled, err := strconv.ParseBool(enabledStr)
	if err != nil {
		helpers.RenderErrorJSON(w, http.StatusBadRequest, "Invalid enabled value")
		return
	}

	path := fmt.Sprintf("%s/mod-list.json", s.fsmConfig.Factorio.ModsDir)
	err = mods.SetModEnabled(path, modName, enabled)
	if err != nil {
		log.Printf("Failed to update %s: %v", path, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to update mod list")
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Failed to read %s: %v", path, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to read mod list")
		return
	}
	var modList mods.ModList
	if err := json.Unmarshal(data, &modList); err != nil {
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to parse mod list")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(modList)
}
