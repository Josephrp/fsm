// Package server provides HTTP handler functions for managing FSM admin users,
// including listing, adding, updating, and deleting admins.
package server

import (
	"encoding/json"
	"html"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/snarf-dev/fsm/v2/internal/auth"
	"github.com/snarf-dev/fsm/v2/internal/helpers"
	"github.com/snarf-dev/fsm/v2/internal/validators"
)

// handleAddAdmin creates a new admin user with a hashed password and saves it to the config.
// Expects a JSON body with "username" and "password" fields.
func (s *RestServer) handleAddAdmin(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		helpers.RenderErrorJSON(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if !validators.IsUsernameValid(payload.Username) {
		helpers.RenderErrorJSON(w, http.StatusBadRequest, "Invalid username")
		return
	}

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Unable to hash password")
		return
	}
	s.fsmConfig.Admins[payload.Username] = hashedPassword
	s.fsmConfig.SaveToFile()
	w.WriteHeader(http.StatusNoContent)
}

// handleUpdateAdmin updates the password for an existing admin user.
// Expects a JSON body with a "password" field and the username in the route variable.
func (s *RestServer) handleUpdateAdmin(w http.ResponseWriter, r *http.Request) {
	user := mux.Vars(r)["user"]
	var payload struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		helpers.RenderErrorJSON(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if !validators.IsUsernameValid(user) {
		helpers.RenderErrorJSON(w, http.StatusBadRequest, "Invalid username")
		return
	}

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		helpers.RenderErrorJSON(w, http.StatusInternalServerError, "Unable to hash password")
		return
	}
	s.fsmConfig.Admins[user] = hashedPassword
	s.fsmConfig.SaveToFile()
	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteAdmin removes an admin user from the config unless the user is deleting themselves.
// The username to delete is provided in the route variable.
func (s *RestServer) handleDeleteAdmin(w http.ResponseWriter, r *http.Request) {
	user := mux.Vars(r)["user"]

	if !validators.IsUsernameValid(user) {
		helpers.RenderErrorJSON(w, http.StatusBadRequest, "Invalid username")
		return
	}

	username, _, _ := r.BasicAuth()
	if user == username {
		helpers.RenderErrorJSON(w, http.StatusForbidden, "Cannot delete yourself")
		return
	}
	delete(s.fsmConfig.Admins, user)
	s.fsmConfig.SaveToFile()
	w.WriteHeader(http.StatusNoContent)
}

// handleListAdmins returns a list of all admin usernames with password fields redacted.
func (s *RestServer) handleListAdmins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	admins := make(map[string]string)
	for k := range s.fsmConfig.Admins {
		admins[html.EscapeString(k)] = ""
	}
	json.NewEncoder(w).Encode(admins)
}
