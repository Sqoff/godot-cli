//go:build !windows

package instance

import (
	"os"
	"syscall"
)

func pidExists(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal(0) checks existence without sending a real signal
	return p.Signal(syscall.Signal(0)) == nil
}
