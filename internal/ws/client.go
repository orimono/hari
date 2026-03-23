package ws

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/orimono/hari/internal/config"
	"github.com/orimono/hari/internal/dispatcher"
	"github.com/orimono/hari/internal/logger"
	"github.com/orimono/hari/internal/protocol"
)

type Client struct {
	upgrader websocket.Upgrader
	cfg      *config.Config
}

var GlobalClient *Client

type Session struct {
	conn      *websocket.Conn
	send      chan []byte
	done      chan struct{}
	closeOnce sync.Once
}

func newClient() *Client {
	cfg := config.MustLoad()
	logger.Init(cfg.LogLevel)
	return &Client{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		cfg: cfg,
	}
}

func Serve() {
	GlobalClient = newClient()
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(GlobalClient.cfg.ServerURL, nil)
	if err != nil {
		slog.Error("Failed to connect to server", "error", err)
		return
	}
	session := &Session{
		conn: conn,
		send: make(chan []byte),
		done: make(chan struct{}),
	}
	session.run()
}

func (s *Session) run() {
	go s.setHeartbeat()
	go s.readerLoop()
	s.writerLoop()
}

func (s *Session) readerLoop() {
	defer close(s.done)
	for {
		_, data, err := s.conn.ReadMessage()
		if err != nil {
			slog.Warn("Connection lost", "error", err)
			return
		}

		msg, err := protocol.Decode(data)
		if err != nil {
			slog.Error("Failed to decode message from data", "error", err)
		}

		dispatcher.Dispatch(msg)
	}
}

func (s *Session) writerLoop() {

}

func (s *Session) setHeartbeat() {
	s.conn.SetReadDeadline(time.Now().Add(time.Duration(GlobalClient.cfg.PongTimeout)))
	s.conn.SetPongHandler(func(string) error {
		return s.conn.SetReadDeadline(
			time.Now().Add(time.Duration(GlobalClient.cfg.PongTimeout)))
	})
}
