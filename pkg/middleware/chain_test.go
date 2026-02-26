package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChain_Order(t *testing.T) {
	t.Parallel()

	var got []string

	mw := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = append(got, name+"-before")
				next.ServeHTTP(w, r)
				got = append(got, name+"-after")
			})
		}
	}

	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got = append(got, "handler")
			w.WriteHeader(204)
		}),
		mw("A"),
		mw("B"),
	)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != 204 {
		t.Fatalf("status=%d", rr.Code)
	}

	want := []string{
		"A-before",
		"B-before",
		"handler",
		"B-after",
		"A-after",
	}

	if len(got) != len(want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got=%v want=%v", got, want)
		}
	}
}

func TestChain_NoMiddleware(t *testing.T) {
	t.Parallel()

	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != 204 {
		t.Fatalf("status=%d", rr.Code)
	}
}
