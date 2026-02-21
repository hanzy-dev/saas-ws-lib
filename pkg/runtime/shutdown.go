package runtime

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Shutdown struct {
	Timeout time.Duration
	hooks   []func(context.Context) error
}

func New(timeout time.Duration) *Shutdown {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Shutdown{Timeout: timeout}
}

func (s *Shutdown) Add(fn func(context.Context) error) {
	if fn == nil {
		return
	}
	s.hooks = append(s.hooks, fn)
}

// Wait blocks until SIGINT/SIGTERM is received, then runs hooks in reverse order.
// Returns the first error encountered (but still attempts all hooks).
func (s *Shutdown) Wait(ctx context.Context) error {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(ch)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
	}

	cctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	var firstErr error
	for i := len(s.hooks) - 1; i >= 0; i-- {
		if err := s.hooks[i](cctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
