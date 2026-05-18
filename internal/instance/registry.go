package instance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Instance represents a running Godot editor registered in the file registry.
type Instance struct {
	PID           int    `json:"pid"`
	Port          int    `json:"port"`
	ProjectPath   string `json:"project_path"`
	GodotVersion  string `json:"godot_version"`
	Token         string `json:"token"`
	StartedAt     string `json:"started_at"`
	LastHeartbeat string `json:"last_heartbeat"`
}

// Dir returns the path to the instances directory (~/.godot-cli/instances).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".godot-cli", "instances"), nil
}

// List returns all valid instance files. Malformed files are silently skipped.
func List() ([]*Instance, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var out []*Instance
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		inst, err := readFile(filepath.Join(dir, e.Name()))
		if err == nil {
			out = append(out, inst)
		}
	}
	return out, nil
}

func readFile(path string) (*Instance, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var inst Instance
	return &inst, json.Unmarshal(data, &inst)
}
