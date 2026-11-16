package ingest

import "context"

// MessageType represents a WebSocket message type.
type MessageType int

const (
	// MessageText indicates a UTF-8 text message.
	MessageText MessageType = iota + 1
	// MessageBinary indicates a binary message.
	MessageBinary
)

// Message is a single WebSocket message payload.
type Message struct {
	Type MessageType
	Data []byte
}

// Conn is the minimal WebSocket connection contract needed by the ingester.
type Conn interface {
	Read(ctx context.Context) (Message, error)
	Close(code int, reason string) error
}

// Dialer establishes WebSocket connections.
type Dialer interface {
	Dial(ctx context.Context, url string) (Conn, error)
}
