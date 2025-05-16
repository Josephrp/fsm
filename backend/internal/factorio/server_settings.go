package factorio

// Package factorio provides utilities for interacting with the Factorio server
// configuration and settings, including loading and managing server settings files.

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ReadServerSettings reads a Factorio server-settings.json file and parses it
// into a map of key-value pairs. It returns the parsed configuration or an error.
func ReadServerSettings(settings_file string) (map[string]interface{}, error) {
	if data, err := os.ReadFile(filepath.Clean(settings_file)); err != nil {
		return nil, err
	} else {
		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err != nil {
			return nil, err
		}

		return settings, nil
	}
}
