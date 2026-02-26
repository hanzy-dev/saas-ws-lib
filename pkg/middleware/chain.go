package middleware

import "net/http"

// Middleware represents an HTTP middleware function.
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares to handler.
// Middlewares are applied in the order provided:
// Chain(h, A, B) => A(B(h))
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
