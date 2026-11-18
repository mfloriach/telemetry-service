package ws

import (
	"context"

	"github.com/coder/websocket"
)

type Conn interface {
	Read(ctx context.Context) (websocket.MessageType, []byte, error)
	Close(code int, reason string) error
}

type CoderDialer struct{}

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

func (c *coderConn) Read(ctx context.Context) (websocket.MessageType, []byte, error) {
	return c.conn.Read(ctx)
}

func (c *coderConn) Close(code int, reason string) error {
	return c.conn.Close(websocket.StatusCode(code), reason)
}
