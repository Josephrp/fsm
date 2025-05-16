package server

// Package server contains HTTP handler logic for the Factorio Server Manager,
// including endpoints for interacting with Factorio's ban list.

import (
	"net/http"

	"github.com/snarf-dev/fsm/v2/internal/helpers"
)

// handleListFactorioBans returns the list of Factorio server bans as JSON.
func (s *RestServer) handleListFactorioBans(w http.ResponseWriter, r *http.Request) {
	helpers.HandleListUsernameFile(s.fsmConfig.Factorio.Files.BanList, w, r)
}

// handleAddFactorioBan adds a new username to the Factorio ban list if not already present.
// It expects a JSON payload with a "username" field.
func (s *RestServer) handleAddFactorioBanUser(w http.ResponseWriter, r *http.Request) {
	helpers.HandleAddUsernameToFile(s.fsmConfig.Factorio.Files.BanList, w, r)
}

// handleRemoveFactorioBan removes the specified user from the Factorio ban list.
// The username is taken from the URL path parameter.
func (s *RestServer) handleRemoveFactorioBanUser(w http.ResponseWriter, r *http.Request) {
	helpers.HandleRemoveUsernameFromFile(s.fsmConfig.Factorio.Files.BanList, w, r)
}
