package ws

import (
	"context"
	"log/slog"

	"github.com/gorilla/websocket"
	"github.com/orimono/hari/internal/config"
	"github.com/orimono/hari/internal/protocol"
)

type Client struct {
	cfg     *config.Config
	session *Session
	ready   chan *Session
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg:   cfg,
		ready: make(chan *Session, 1),
	}
}

func (c *Client) Serve(ctx context.Context) {
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(c.cfg.ServerURL, nil)
	if err != nil {
		slog.Error("Failed to connect to server", "error", err)
		return
	}
	c.session = NewSession(conn, c.cfg)
	c.ready <- c.session
	c.session.run(ctx)
}

func (c *Client) Send(data []byte) error {
	session := <-c.ready
	session.send <- protocol.Message{
		Type: websocket.TextMessage,
		Data: data,
	}
	return nil
}

func (c *Client) Ready() <-chan *Session {
	return c.ready
}
