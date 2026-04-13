package protocol

// Message wraps a raw WebSocket frame.
type Message struct {
	Type int
	Data []byte
}
