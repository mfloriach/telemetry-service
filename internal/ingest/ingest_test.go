package ingest

import (
	"context"
	"errors"
	"server/internal/utils/ws"
)

type fakeDialer struct {
	conn ws.Conn
	err  error
	url  string
}

func (f *fakeDialer) Dial(_ context.Context, url string) (ws.Conn, error) {
	f.url = url
	if f.err != nil {
		return nil, f.err
	}
	return f.conn, nil
}

type fakeConn struct {
	msgs []ws.Message
	i    int
}

func (f *fakeConn) Read(ctx context.Context) (ws.Message, error) {
	if f.i >= len(f.msgs) {
		return ws.Message{}, errors.New("stream closed")
	}
	m := f.msgs[f.i]
	f.i++
	return m, nil
}

func (f *fakeConn) Close(int, string) error { return nil }
