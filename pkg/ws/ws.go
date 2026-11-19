package ws

import (
	"context"

	"github.com/coder/websocket"
)

type CoderDialer struct{}

func (CoderDialer) Dial(ctx context.Context, url string) (*coderConn, error) {
	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return nil, err
	}
	return &coderConn{conn: c}, nil
}

type coderConn struct {
	conn *websocket.Conn
}

func (c *coderConn) Read(ctx context.Context) (string, []byte, error) {
	message_type, msg, err := c.conn.Read(ctx)
	return message_type.String(), msg, err
}

func (c *coderConn) Close(code int, reason string) error {
	return c.conn.Close(websocket.StatusCode(code), reason)
}
