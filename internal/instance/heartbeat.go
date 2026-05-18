package instance

import "time"

const StaleDuration = 30 * time.Second

// IsStale reports whether the instance's last heartbeat is older than 30 seconds.
func (inst *Instance) IsStale() bool {
	if inst.LastHeartbeat == "" {
		return true
	}
	// Godot emits "2006-01-02T15:04:05" (no timezone suffix)
	t, err := time.Parse("2006-01-02T15:04:05", inst.LastHeartbeat)
	if err != nil {
		return true
	}
	return time.Since(t) > StaleDuration
}
