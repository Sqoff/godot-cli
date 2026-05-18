package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPTransport struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewHTTPTransport(host string, port int, token string) *HTTPTransport {
	return &HTTPTransport{
		baseURL: fmt.Sprintf("http://%s:%d", host, port),
		token:   token,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *HTTPTransport) Send(ctx context.Context, command string, params map[string]any) (*Response, error) {
	if params == nil {
		params = map[string]any{}
	}
	body, _ := json.Marshal(params)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/api/"+command, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")

	resp, err := t.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

func (t *HTTPTransport) Ping(ctx context.Context) error {
	resp, err := t.Send(ctx, "ping", nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		if resp.Error != nil {
			return fmt.Errorf("ping: %s", resp.Error.Message)
		}
		return fmt.Errorf("ping: unknown error")
	}
	return nil
}

func (t *HTTPTransport) Close() error { return nil }
