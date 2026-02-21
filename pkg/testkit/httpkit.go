package testkit

import (
	"net/http"
	"net/http/httptest"

	wslog "saas-ws-lib/pkg/log"
	"saas-ws-lib/pkg/middleware"
)

type HTTPTestServer struct {
	Server *httptest.Server
	Client *http.Client
}

type HTTPServerOptions struct {
	Logger            *wslog.Logger
	BodyLimit         int64
	MetricsRouteLabel string // optional
}

// NewServer spins up an httptest server with a standard middleware chain.
// Keep it router-agnostic: caller provides the final handler.
func NewServer(h http.Handler, opts HTTPServerOptions) *HTTPTestServer {
	if opts.Logger == nil {
		opts.Logger = wslog.NewJSON(wslog.Options{}) // defaults to stdout; acceptable for tests
	}
	if opts.BodyLimit == 0 {
		opts.BodyLimit = 1 << 20 // 1 MiB
	}

	chain := middleware.Chain(
		h,
		middleware.RequestID(),
		middleware.Recover(opts.Logger),
		middleware.BodyLimit(opts.BodyLimit),
	)

	srv := httptest.NewServer(chain)
	return &HTTPTestServer{
		Server: srv,
		Client: srv.Client(),
	}
}

func (s *HTTPTestServer) Close() {
	if s != nil && s.Server != nil {
		s.Server.Close()
	}
}
