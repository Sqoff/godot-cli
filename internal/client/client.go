package client

import "context"

type Client struct {
	t Transport
}

func New(transport Transport) *Client {
	return &Client{t: transport}
}

func (c *Client) Send(ctx context.Context, command string, params map[string]any) (*Response, error) {
	return c.t.Send(ctx, command, params)
}

func (c *Client) Ping(ctx context.Context) error {
	return c.t.Ping(ctx)
}

func (c *Client) Close() error {
	return c.t.Close()
}
