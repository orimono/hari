package ws

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/orimono/shutter/internal/config"
	"github.com/orimono/shutter/internal/dispatcher"
	"github.com/orimono/shutter/internal/protocol"
	"github.com/orimono/shutter/internal/util"
)

var ErrNoSession = errors.New("no active session")

type Client struct {
	cfg        *config.Config
	dispatcher *dispatcher.Dispatcher
	mu         sync.RWMutex
	session    *Session
	ready      chan struct{}
}

func NewClient(cfg *config.Config, d *dispatcher.Dispatcher) *Client {
	return &Client{
		cfg:        cfg,
		dispatcher: d,
		ready:      make(chan struct{}, 1),
	}
}

func (c *Client) Run(ctx context.Context) {
	dialer := websocket.DefaultDialer

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		util.DrainChan(c.ready)
		conn, _, err := dialer.Dial(c.cfg.ServerURL, nil)
		if err != nil {
			slog.Error("Failed to connect to server", "error", err)
			time.Sleep(time.Duration(c.cfg.RetryInterval))
			continue
		}

		s := NewSession(conn, c.cfg, c.dispatcher)
		c.mu.Lock()
		c.session = s
		c.mu.Unlock()

		c.ready <- struct{}{}
		s.run(ctx)

		c.mu.Lock()
		c.session = nil
		c.mu.Unlock()
	}
}

func (c *Client) Send(data []byte) error {
	c.mu.RLock()
	session := c.session
	c.mu.RUnlock()

	if session == nil {
		return fmt.Errorf("%w", ErrNoSession)
	}
	session.send <- protocol.Message{
		Type: websocket.TextMessage,
		Data: data,
	}
	return nil
}

func (c *Client) Ready() <-chan struct{} {
	return c.ready
}
