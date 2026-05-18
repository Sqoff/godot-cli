package cmdutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sqoff/godot-cli/internal/instance"
)

// Resolve finds the best matching Godot editor instance.
//
// Priority: --port flag > --project flag > current directory > only one instance.
func Resolve(flagPort int, flagProject string) (*instance.Instance, error) {
	if _, err := instance.CleanZombies(); err != nil {
		fmt.Fprintf(os.Stderr, "warn: zombie cleanup: %v\n", err)
	}

	instances, err := instance.List()
	if err != nil {
		return nil, fmt.Errorf("cannot read instances: %w", err)
	}
	if len(instances) == 0 {
		return nil, fmt.Errorf("no active Godot editor instance found")
	}

	// --port override
	if flagPort > 0 {
		for _, inst := range instances {
			if inst.Port == flagPort {
				return inst, nil
			}
		}
		return nil, fmt.Errorf("no instance found on port %d", flagPort)
	}

	// --project override
	if flagProject != "" {
		abs, _ := filepath.Abs(flagProject)
		for _, inst := range instances {
			if inst.ProjectPath == abs {
				return inst, nil
			}
		}
	}

	// Match current working directory
	if cwd, err := os.Getwd(); err == nil {
		for _, inst := range instances {
			if inst.ProjectPath == cwd {
				return inst, nil
			}
		}
	}

	// Single instance: use it without ambiguity
	if len(instances) == 1 {
		return instances[0], nil
	}

	return nil, fmt.Errorf("multiple instances found; use --port to specify one")
}
