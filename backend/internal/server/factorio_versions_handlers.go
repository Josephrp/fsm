package server

// Package server provides HTTP handlers for managing Factorio server versions,
// including downloading, selecting, uninstalling, and monitoring download progress.

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/snarf-dev/fsm/v2/internal/factorio"
	"github.com/snarf-dev/fsm/v2/internal/helpers"
)

// handleDownloadFactorioVersion triggers download and extraction of a specified Factorio version.
// Expects `branch` and `version` path parameters.
func (s *RestServer) handleDownloadFactorioVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branch := vars["branch"]
	version := vars["version"]

	_, err := factorio.DownloadAndExtractVersion(s.fsmConfig, branch, version)
	if err != nil {
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to read API response")
		return
	}
}

// handleDownloadProgressStream establishes a WebSocket connection that streams
// download and extraction progress updates for a specific Factorio version.
func (s *RestServer) handleDownloadProgressStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branch := vars["branch"]
	version := vars["version"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}
	defer conn.Close()

	progressCh := factorio.SubscribeDownloadProgress(branch, version)

	for progress := range progressCh {
		err := conn.WriteJSON(map[string]interface{}{
			"type":    "progress",
			"branch":  branch,
			"version": version,
			"percent": progress.Percent,
			"stage":   progress.Stage,
		})
		if err != nil {
			log.Printf("WebSocket write failed: %v\n", err)
			break
		}
	}
}

// handleListFactorioVersions returns a combined response with available and installed Factorio versions.
func (s *RestServer) handleListFactorioVersions(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("https://factorio.com/api/latest-releases")
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Error talking to Factorio server: %v\n", err)
		helpers.RenderErrorJSON(w, http.StatusBadGateway, "Failed to query Factorio versions API")
		return
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to read API response")
		return
	}

	installed, _ := factorio.GetInstalledFactorioVersions(s.fsmConfig.Factorio.ServerVersions)

	response := map[string]interface{}{
		"available": json.RawMessage(buf.Bytes()),
		"installed": installed,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSelectFactorioVersion updates the configuration to use a specified Factorio version.
// Expects `branch` and `version` path parameters.
func (s *RestServer) handleSelectFactorioVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branch := vars["branch"]
	version := vars["version"]

	err := factorio.SelectVersion(s.fsmConfig, branch, version)
	if err != nil {
		log.Printf("Failed to switch version:%v\n", err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to switch versions")
		return
	}
}

// handleUninstallFactorioVersion removes the files for the specified Factorio version.
// Expects `branch` and `version` path parameters.
func (s *RestServer) handleUninstallFactorioVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branch := vars["branch"]
	version := vars["version"]

	err := factorio.UninstallVersion(s.fsmConfig, branch, version)
	if err != nil {
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to uninstall")
		return
	}
}
