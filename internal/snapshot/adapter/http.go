package adapter

import (
	"context"
	"errors"
	"io"
	"net/http"
)

type Handler struct{ Next http.Handler }

func ServeSnapshotExport(ctx context.Context, open func(context.Context) (io.ReadCloser, error), copyBody func(io.Reader) error) (err error) {
	reader, err := open(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	return copyBody(reader)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Next != nil {
		h.Next.ServeHTTP(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func MethodAllowed(method string) bool { return method == http.MethodGet || method == http.MethodPost }
