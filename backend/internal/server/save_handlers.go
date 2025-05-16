package server

// Package server provides RESTful HTTP handlers for managing Factorio save files,
// including listing, uploading, downloading, and deleting saves.

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/snarf-dev/fsm/v2/internal/helpers"
)

// handleListSaves returns a JSON list of all save files in the configured saves directory.
// Each entry includes the name, size, and last modified time.
func (s *RestServer) handleListSaves(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(s.fsmConfig.Factorio.SavesDir)
	if err != nil {
		log.Printf("Failed to read %s: %v", s.fsmConfig.Factorio.SavesDir, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to list saves")
		return
	}
	type Save struct {
		Name    string    `json:"name"`
		Size    int64     `json:"size"`
		ModTime time.Time `json:"modTime"`
	}
	var saves []Save
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		info, err := f.Info()
		if err != nil {
			continue
		}
		saves = append(saves, Save{
			Name:    f.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(saves)
}

// handleDownloadSave streams a specified save file to the client as a file download.
// The file name is passed as a URL path variable.
func (s *RestServer) handleDownloadSave(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	filePath := filepath.Join(s.fsmConfig.Factorio.SavesDir, filepath.Clean(name))
	w.Header().Set("Content-Disposition", "attachment; filename="+name)
	http.ServeFile(w, r, filePath)
}

// handleUploadSave accepts a multipart file upload from a form and writes it to the saves directory.
// The file is expected to be sent under the "save" form field.
func (s *RestServer) handleUploadSave(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Printf("Error: %v\n", err)
		helpers.RenderErrorJSON(w, http.StatusBadRequest, "Failed to parse form")
		return
	}
	file, header, err := r.FormFile("save")
	if err != nil {
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to get uploaded file")
		return
	}
	defer file.Close()

	destPath := filepath.Join(s.fsmConfig.Factorio.SavesDir, filepath.Base(header.Filename))
	out, err := os.Create(destPath)
	if err != nil {
		log.Printf("Failed to write %s: %v", destPath, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to create file")
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		log.Printf("Failed to write %s: %v", destPath, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to write file")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// handleDeleteSave removes a specified save file from the saves directory.
// The file name is passed as a URL path variable.
func (s *RestServer) handleDeleteSave(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	filePath := filepath.Join(s.fsmConfig.Factorio.SavesDir, filepath.Clean(name))
	err := os.Remove(filePath)
	if err != nil {
		log.Printf("Failed to delete %s: %v", filePath, err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Failed to delete save")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
