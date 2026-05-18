package instance

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// CleanZombies removes instance files whose PID no longer exists on the OS.
// Returns the number of files removed.
func CleanZombies() (int, error) {
	dir, err := Dir()
	if err != nil {
		return 0, err
	}

	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	removed := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		pidStr := strings.TrimSuffix(e.Name(), ".json")
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}
		if !pidExists(pid) {
			if os.Remove(filepath.Join(dir, e.Name())) == nil {
				removed++
			}
		}
	}
	return removed, nil
}
