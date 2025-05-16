package server

// Package server contains HTTP handler logic for the Factorio Server Manager,
// including endpoints for interacting with Factorio's white list.

import (
	"net/http"

	"github.com/snarf-dev/fsm/v2/internal/helpers"
)

// handleListFactorioWhitelistUsers returns the list of Factorio server white listed users as JSON.
func (s *RestServer) handleListFactorioWhitelistUsers(w http.ResponseWriter, r *http.Request) {
	helpers.HandleListUsernameFile(s.fsmConfig.Factorio.Files.WhiteList, w, r)
}

// handleAddFactorioWhitelistUser adds a new username to the Factorio white list if not already present.
// It expects a JSON payload with a "username" field.
func (s *RestServer) handleAddFactorioWhitelistUser(w http.ResponseWriter, r *http.Request) {
	helpers.HandleAddUsernameToFile(s.fsmConfig.Factorio.Files.WhiteList, w, r)
}

// handleRemoveFactorioWhitelistUser removes the specified user from the Factorio white list.
// The username is taken from the URL path parameter.
func (s *RestServer) handleRemoveFactorioWhitelistUser(w http.ResponseWriter, r *http.Request) {
	helpers.HandleRemoveUsernameFromFile(s.fsmConfig.Factorio.Files.WhiteList, w, r)
}
