package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/gin-contrib/sse"
)

// runapi starts http API server.
// Cancelling ctx will shutdown the http server gracefully.
func runapi(ctx context.Context, addr string, mcpsrv *MCPServer) error {
	sessions := NewSessions()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ack", func(_ http.ResponseWriter, _ *http.Request) {})

	mux.HandleFunc("GET /sse", cors(handleSSE(ctx, sessions, mcpsrv)))
	mux.HandleFunc("POST /messages", cors(handleMessage(ctx, sessions, mcpsrv)))
	mux.HandleFunc("OPTIONS /messages", cors(handleMessage(ctx, sessions, mcpsrv)))

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

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("access-control-allow-credentials", "true")
		w.Header().Set("access-control-allow-origin", "*")

		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		if req.Method == http.MethodOptions {
			w.Header().Set("access-control-allow-methods", "GET,POST")
			w.Header().Set("access-control-allow-headers", "content-type,Accept,Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, req)
	}
}

func handleSSE(_ context.Context, sessions *Sessions, srv *MCPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		w.Header().Set("Content-Type", "text/event-stream")

		s := NewSession()
		defer s.Close()

		sessions.Add(s)
		defer sessions.Remove(s.id)

		slog.Debug("connect sse", slog.Any("id", s.id))
		defer slog.Debug("disconnect sse", slog.Any("id", s.id))

		// create the mcpconn
		mcpconn := srv.NewConn()
		defer mcpconn.Close()

		f, ok := w.(http.Flusher)
		if !ok {
			panic("response writer not a flusher")
		}

		send := func(event string, data any) error {
			err := sse.Encode(w, sse.Event{
				Event: event,
				Data:  data,
			})
			if err != nil {
				return fmt.Errorf("encode: %s", err)
			}
			f.Flush()
			return nil
		}

		if err := send("endpoint", fmt.Sprintf("/messages?id=%s", s.id)); err != nil {
			return
		}

		for {
			select {
			case rreq, ok := <-s.Requests():
				if !ok {
					// closed channel
					return
				}
				srv.Handle(ctx, rreq, mcpconn, send)
			case <-req.Context().Done():
				return
			case <-ctx.Done():
				return
			}
		}

	}
}

func handleMessage(_ context.Context, sessions *Sessions, srv *MCPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// get the sessionId
		var id SessionId
		if err := id.Set(req.URL.Query().Get("id")); err != nil {
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}

		// retrieve the session
		s, ok := sessions.Get(SessionId(id))
		if !ok {
			slog.Debug("invalid session id", slog.Any("id", id))
			http.Error(w, "id not found", http.StatusBadRequest)
			return
		}

		mcpreq, err := srv.Decode(req.Body)
		if err != nil {
			slog.Error("message decode error", slog.Any("err", err))
			if errors.Is(err, ErrRPCRequest) {
				// TODO disconnect the client?
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte("Accepted"))
				return
			}

			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		s.Requests() <- mcpreq

		w.WriteHeader(http.StatusAccepted)
	}
}
