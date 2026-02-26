package httpx

import (
	"net/http"
	"testing"
	"time"
)

func TestNewServer_DefaultTimeouts(t *testing.T) {
	t.Parallel()

	srv := NewServer(ServerConfig{Addr: ":0"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	if srv.Addr != ":0" {
		t.Fatalf("addr=%q", srv.Addr)
	}
	if srv.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout=%s", srv.ReadHeaderTimeout)
	}
	if srv.ReadTimeout != 15*time.Second {
		t.Fatalf("ReadTimeout=%s", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 15*time.Second {
		t.Fatalf("WriteTimeout=%s", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout=%s", srv.IdleTimeout)
	}
}

func TestNewServer_CustomTimeouts(t *testing.T) {
	t.Parallel()

	cfg := ServerConfig{
		Addr:              "127.0.0.1:1234",
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       4 * time.Second,
	}
	srv := NewServer(cfg, nil)

	if srv.Addr != cfg.Addr {
		t.Fatalf("addr=%q", srv.Addr)
	}
	if srv.ReadHeaderTimeout != cfg.ReadHeaderTimeout {
		t.Fatalf("ReadHeaderTimeout=%s", srv.ReadHeaderTimeout)
	}
	if srv.ReadTimeout != cfg.ReadTimeout {
		t.Fatalf("ReadTimeout=%s", srv.ReadTimeout)
	}
	if srv.WriteTimeout != cfg.WriteTimeout {
		t.Fatalf("WriteTimeout=%s", srv.WriteTimeout)
	}
	if srv.IdleTimeout != cfg.IdleTimeout {
		t.Fatalf("IdleTimeout=%s", srv.IdleTimeout)
	}
	if srv.Handler != nil {
		t.Fatalf("expected nil handler passthrough")
	}
}
