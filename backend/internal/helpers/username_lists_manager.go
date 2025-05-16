package helpers

import (
	"encoding/json"
	"html"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/snarf-dev/fsm/v2/internal/validators"
)

func HandleListUsernameFile(path string, w http.ResponseWriter, r *http.Request) {
	fileData, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Failed to read %s: %v", path, err)
		RenderErrorJSON(w, http.StatusInternalServerError, "Could not read admin list")
		return
	}

	var usernames []string
	if err := json.Unmarshal(fileData, &usernames); err != nil {
		log.Printf("Invalid JSON in %s: %v", path, err)
		RenderErrorJSON(w, http.StatusInternalServerError, "Invalid user list format")
		return
	}

	sanitized := make([]string, 0, len(usernames))
	for _, username := range usernames {
		sanitized = append(sanitized, html.EscapeString(username))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sanitized); err != nil {
		log.Printf("Failed to encode sanitized user list: %v", err)
		RenderErrorJSON(w, http.StatusInternalServerError, "Failed to encode response")
	}
}

func HandleAddUsernameToFile(path string, w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Username == "" {
		RenderErrorJSON(w, http.StatusBadRequest, "Invalid payload")
		return
	}

	if !validators.IsUsernameValid(payload.Username) {
		RenderErrorJSON(w, http.StatusBadRequest, "Invalid username")
		return
	}

	var admins []string

	if file, err := os.ReadFile(path); err == nil {
		json.Unmarshal(file, &admins)
	}

	for _, name := range admins {
		if strings.EqualFold(name, payload.Username) {
			RenderErrorJSON(w, http.StatusConflict, "User already in admin list")
			return
		}
	}

	admins = append(admins, payload.Username)
	data, _ := json.MarshalIndent(admins, "", "  ")
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("Failed to write %s: %v", path, err)
		RenderErrorJSON(w, http.StatusInternalServerError, "Failed to write admin list")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func HandleRemoveUsernameFromFile(path string, w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["user"]
	if username == "" {
		RenderErrorJSON(w, http.StatusBadRequest, "Missing username")
		return
	}

	if !validators.IsUsernameValid(username) {
		RenderErrorJSON(w, http.StatusBadRequest, "Invalid username")
		return
	}

	var admins []string
	if file, err := os.ReadFile(path); err == nil {
		json.Unmarshal(file, &admins)
	}

	newAdmins := make([]string, 0, len(admins))
	for _, name := range admins {
		if !strings.EqualFold(name, username) {
			newAdmins = append(newAdmins, name)
		}
	}

	data, _ := json.MarshalIndent(newAdmins, "", "  ")
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("Failed to write %s: %v", path, err)
		RenderErrorJSON(w, http.StatusInternalServerError, "Failed to write admin list")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
