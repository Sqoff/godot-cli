//go:build windows

package instance

import "syscall"

func pidExists(pid int) bool {
	// PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	h, err := syscall.OpenProcess(0x1000, false, uint32(pid))
	if err != nil {
		return false
	}
	_ = syscall.CloseHandle(h)
	return true
}
