package ws

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/orimono/shutter/internal/config"
	"github.com/orimono/shutter/internal/dispatcher"
	"github.com/orimono/shutter/internal/protocol"
)

type Session struct {
	conn       *websocket.Conn
	cfg        *config.Config
	dispatcher *dispatcher.Dispatcher
	send       chan protocol.Message
	done       chan struct{}
	closeOnce  sync.Once
}

func NewSession(conn *websocket.Conn, cfg *config.Config, d *dispatcher.Dispatcher) *Session {
	return &Session{
		conn:       conn,
		cfg:        cfg,
		dispatcher: d,
		send:       make(chan protocol.Message, 100),
		done:       make(chan struct{}),
	}
}

func (s *Session) run(ctx context.Context) {
	go s.readerLoop(ctx)
	go s.writerLoop(ctx)
	go s.setHeartbeat(ctx)

	select {
	case <-ctx.Done():
	case <-s.done:
	}
	s.conn.Close()
}

func (s *Session) readerLoop(ctx context.Context) {
	defer close(s.done)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, data, err := s.conn.ReadMessage()
		if err != nil {
			slog.Warn("read failed", "error", err)
			return
		}

		s.dispatcher.Handle(ctx, data, func(msg protocol.Message) {
			s.send <- msg
		})
	}
}

func (s *Session) writerLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case msg := <-s.send:
			s.conn.SetWriteDeadline(time.Now().Add(time.Duration(s.cfg.WriterTimeout)))
			if err := s.conn.WriteMessage(msg.Type, msg.Data); err != nil {
				slog.Error("write failed", "error", err)
				return
			}
		}
	}
}

func (s *Session) setHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(s.cfg.PingInterval))
	defer ticker.Stop()

	s.conn.SetReadDeadline(time.Now().Add(time.Duration(s.cfg.PongTimeout)))
	s.conn.SetPongHandler(func(string) error {
		return s.conn.SetReadDeadline(
			time.Now().Add(time.Duration(s.cfg.PongTimeout)))
	})

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.send <- protocol.Message{
				Type: websocket.PingMessage,
				Data: nil,
			}
		}
	}
}
