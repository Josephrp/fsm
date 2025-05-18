package server

// Package server provides HTTP handlers for interacting with the Factorio mod list,
// allowing users to retrieve and toggle mod states through the REST API.

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/snarf-dev/fsm/v2/internal/factorio"
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

func (s *RestServer) bookmarkedModsHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(fmt.Sprintf("https://mods.factorio.com/api/bookmarks?username=%s&token=%s", s.fsmConfig.Factorio.Username, s.fsmConfig.Factorio.Token))
	if err != nil {
		log.Printf("Error talking to Factorio server: %v\n", err)
		helpers.RenderErrorJSON(w, http.StatusBadGateway, "Failed to query Factorio Bookmarks API")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Factorio API returned status: %d\n", resp.StatusCode)
		helpers.RenderErrorJSON(w, http.StatusBadGateway, "Factorio Bookmarks API returned an error")
		return
	}

	var bookmarks []string
	if err := json.NewDecoder(resp.Body).Decode(&bookmarks); err != nil {
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to parse bookmarks")
		return
	}

	modsInfo := make([]*factorio.ModInfo, 0, len(bookmarks))
	for _, modName := range bookmarks {
		modDetails, err := factorio.GetModDetails(modName)
		if err != nil {
			log.Printf("Failed to fetch mod details for %s: %v\n", modName, err)
			continue // Optionally skip or return error
		}
		sort.Slice(modDetails.Releases, func(i, j int) bool {
			return modDetails.Releases[i].ReleasedAt > modDetails.Releases[j].ReleasedAt
		})
		modsInfo = append(modsInfo, modDetails)
	}

	available, err := factorio.GetAvailableMods(s.fsmConfig)
	if err != nil {
		log.Printf("Failed to get available mods %v\n", err)
		available = []map[string][]string{}
	}

	installed, err := factorio.GetInstalledMods(s.fsmConfig)
	if err != nil {
		log.Printf("Failed to get installed mods %v\n", err)
		installed = []map[string][]string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available":    available,
		"downloadable": modsInfo,
		"installed":    installed,
	})
}

// handleDownloadMod triggers download and extraction of a specified mod version.
// Expects `mod` and `version` path parameters.
func (s *RestServer) handleDownloadMod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mod := vars["mod"]
	version := vars["version"]

	_, err := factorio.DownloadMod(s.fsmConfig, mod, version)
	if err != nil {
		log.Printf("Failed to download mod %s-%s: %v\n", mod, version, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to download mod")
	}
}

// handleInstallMod installs specified mod version into the mods directory.
// Expects `mod` and `version` path parameters.
func (s *RestServer) handleInstallMod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mod := vars["mod"]
	version := vars["version"]

	err := factorio.InstallMod(s.fsmConfig, mod, version)
	if err != nil {
		log.Printf("Failed to install mod %s-%s: %v\n", mod, version, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to install mod")
	}
}

// handleUninstallMod uninstalls specified mod version from the mods directory.
// Expects `mod` and `version` path parameters.
func (s *RestServer) handleUninstallMod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mod := vars["mod"]
	version := vars["version"]

	err := factorio.UninstallMod(s.fsmConfig, mod, version)
	if err != nil {
		log.Printf("Failed to uninstall mod %s-%s: %v\n", mod, version, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to uninstall mod")
	}
}

// handleDeleteMod deletes specified mod version from the downloads directory.
// Expects `mod` and `version` path parameters.
func (s *RestServer) handleDeleteMod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mod := vars["mod"]
	version := vars["version"]

	err := factorio.DeleteMod(s.fsmConfig, mod, version)
	if err != nil {
		log.Printf("Failed to delete mod %s-%s: %v\n", mod, version, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to delete mod")
	}
}
