package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
)

// runapi starts http API server.
// Cancelling ctx will shutdown the http server gracefully.
func runapi(ctx context.Context, addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ack", func(_ http.ResponseWriter, _ *http.Request) {})

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	// shutdown api server on context cancelation
	go func(ctx context.Context, srv *http.Server) {
		<-ctx.Done()
		slog.Debug("api server shutting down")
		// we use context.Background() here b/c ctx is already canceled.
		if err := srv.Shutdown(context.Background()); err != nil {
			// context cancellation error is ignored.
			if !errors.Is(err, context.Canceled) {
				slog.Error("server shutdown", slog.String("err", err.Error()))
			}
		}
	}(ctx, srv)

	slog.Info("server listening", slog.String("addr", addr))

	// ListenAndServe always returns a non-nil error.
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("api server: %w", err)
	}
	slog.Info("api server shutdown")

	return nil
}
