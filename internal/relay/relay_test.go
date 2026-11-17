package relay

import (
	"context"
	"errors"
	"server/internal/utils/ws"
)

type fakeConn struct {
	msgs []ws.Message
	i    int
}

func (c *fakeConn) Read(_ context.Context) (ws.Message, error) {
	if c.i >= len(c.msgs) {
		return ws.Message{}, errors.New("stream closed")
	}
	m := c.msgs[c.i]
	c.i++
	return m, nil
}

func (c *fakeConn) Close(int, string) error { return nil }
