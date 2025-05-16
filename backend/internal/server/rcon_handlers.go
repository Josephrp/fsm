package server

// Package server provides an HTTP interface for sending commands to the Factorio server via RCON.

import (
	"encoding/json"
	"net/http"

	"github.com/gorcon/rcon"
)

// rconHandler processes an HTTP request to send a command via RCON to the Factorio server.
// It returns the command output as JSON or an error message if the command fails.
func (s *RestServer) rconHandler(w http.ResponseWriter, r *http.Request) {
	if !s.fsmConfig.RCon.Enabled {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	command := r.FormValue("command")
	output, err := sendRCONCommand(s.fsmConfig.RCon.Bind, s.fsmConfig.RCon.Password, command)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"output": output,
	})
}

// sendRCONCommand connects to the Factorio RCON interface at the given address using the provided
// password and sends the specified command. It returns the server's response or an error.
func sendRCONCommand(addr, password, command string) (string, error) {
	client, err := rcon.Dial(addr, password)
	if err != nil {
		return "", err
	}
	defer client.Close()

	response, err := client.Execute(command)
	if err != nil {
		return "", err
	}
	return response, nil
}
