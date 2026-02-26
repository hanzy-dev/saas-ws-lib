package errors

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

type errWriter struct {
	h      http.Header
	status int
}

func (w *errWriter) Header() http.Header {
	if w.h == nil {
		w.h = make(http.Header)
	}
	return w.h
}
func (w *errWriter) WriteHeader(code int) { w.status = code }
func (w *errWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestWrite_EncodeFailure_DoesNotPanic(t *testing.T) {
	t.Parallel()

	w := &errWriter{}
	Write(context.Background(), w, 400, InvalidArgument("bad"))
	// pass if no panic; status should be written
	if w.status != 400 {
		t.Fatalf("status=%d want=400", w.status)
	}
}
