package server

// Package server contains HTTP handler logic for the Factorio Server Manager,
// including endpoints for interacting with Factorio's admin list.

import (
	"net/http"

	"github.com/snarf-dev/fsm/v2/internal/helpers"
)

// handleListFactorioAdmins returns the list of Factorio server admins as JSON.
func (s *RestServer) handleListFactorioAdmins(w http.ResponseWriter, r *http.Request) {
	helpers.HandleListUsernameFile(s.fsmConfig.Factorio.Files.AdminList, w, r)
}

// handleAddFactorioAdmin adds a new username to the Factorio admin list if not already present.
// It expects a JSON payload with a "username" field.
func (s *RestServer) handleAddFactorioAdmin(w http.ResponseWriter, r *http.Request) {
	helpers.HandleAddUsernameToFile(s.fsmConfig.Factorio.Files.AdminList, w, r)
}

// handleRemoveFactorioAdmin removes the specified user from the Factorio admin list.
// The username is taken from the URL path parameter.
func (s *RestServer) handleRemoveFactorioAdmin(w http.ResponseWriter, r *http.Request) {
	helpers.HandleRemoveUsernameFromFile(s.fsmConfig.Factorio.Files.AdminList, w, r)
}
