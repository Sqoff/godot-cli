package errors

import "errors"

var (
	ErrNotConnected  = errors.New("no active Godot editor instance found")
	ErrUnauthorized  = errors.New("authentication failed: invalid token")
	ErrCommandFailed = errors.New("command execution failed")
	ErrTimeout       = errors.New("command timed out")
)

const (
	ExitSuccess      = 0
	ExitCommandError = 1
	ExitNoConnection = 2
	ExitUnauthorized = 3
)
