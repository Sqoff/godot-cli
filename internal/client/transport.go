package client

import (
	"context"
	"encoding/json"
)

// Response is the unified JSON response from the Godot plugin.
type Response struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *APIError       `json:"error,omitempty"`
}

// APIError matches the plugin's error schema.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Transport abstracts the HTTP connection to the Godot editor.
// Swapping in a WebSocket transport later requires only a new implementation.
type Transport interface {
	Send(ctx context.Context, command string, params map[string]any) (*Response, error)
	Ping(ctx context.Context) error
	Close() error
}
