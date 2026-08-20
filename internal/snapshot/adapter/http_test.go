package adapter

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

type failingReader struct {
	io.Reader
	closed *bool
	err    error
}

func (r failingReader) Close() error { *r.closed = true; return r.err }

func TestSnapshotHTTPClosesExporter(t *testing.T) {
	copyErr := errors.New("client disconnected")
	closeErr := errors.New("release failed")
	closed := false
	err := ServeSnapshotExport(context.Background(), func(context.Context) (io.ReadCloser, error) {
		return failingReader{Reader: strings.NewReader("snapshot"), closed: &closed, err: closeErr}, nil
	}, func(io.Reader) error { return copyErr })
	if !closed || !errors.Is(err, copyErr) || !errors.Is(err, closeErr) {
		t.Fatalf("export cleanup failed: closed=%v err=%v", closed, err)
	}
}
