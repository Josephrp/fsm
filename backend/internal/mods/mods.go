package mods

// Package mods provides functionality for reading and modifying Factorio mod-list.json files,
// allowing individual mods to be enabled or disabled.

import (
	"encoding/json"
	"os"
)

// ModList represents the structure of a mod-list.json file containing a list of mods and their enabled state.
type ModList struct {
	Mods []ModEntry `json:"mods"`
}

// ModEntry represents a single mod entry with its name and enabled status.
type ModEntry struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// SetModEnabled updates the enabled state of a mod in the given mod-list.json file.
// If the mod does not exist, it is added to the list.
func SetModEnabled(modListPath, modName string, enabled bool) error {
	data, err := os.ReadFile(modListPath)
	if err != nil {
		return err
	}

	var modList ModList
	if err := json.Unmarshal(data, &modList); err != nil {
		return err
	}

	found := false
	for i, mod := range modList.Mods {
		if mod.Name == modName {
			modList.Mods[i].Enabled = enabled
			found = true
			break
		}
	}

	if !found {
		modList.Mods = append(modList.Mods, ModEntry{Name: modName, Enabled: enabled})
	}

	updated, err := json.MarshalIndent(modList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(modListPath, updated, 0644)
}
