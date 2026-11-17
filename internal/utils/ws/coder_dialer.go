package ws

import (
	"context"

	"github.com/coder/websocket"
)

// CoderDialer dials WebSocket connections using github.com/coder/websocket.
type CoderDialer struct{}

// Dial connects to the given WebSocket URL.
func (CoderDialer) Dial(ctx context.Context, url string) (Conn, error) {
	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return nil, err
	}
	return &coderConn{conn: c}, nil
}

type coderConn struct {
	conn *websocket.Conn
}

func (c *coderConn) Read(ctx context.Context) (Message, error) {
	typ, data, err := c.conn.Read(ctx)
	if err != nil {
		return Message{}, err
	}
	msgType := MessageBinary
	if typ == websocket.MessageText {
		msgType = MessageText
	}
	return Message{Type: msgType, Data: data}, nil
}

func (c *coderConn) Close(code int, reason string) error {
	return c.conn.Close(websocket.StatusCode(code), reason)
}
