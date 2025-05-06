package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/lightpanda-io/gomcp/mcp"
)

func runstd(ctx context.Context, in io.Reader, out io.Writer, mcpsrv *MCPServer) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cin := make(chan []byte)
	cout := make(chan mcp.Request)

	enc := json.NewEncoder(out)

	// create the mcpconn
	mcpconn := mcpsrv.NewConn()
	defer mcpconn.Close()

	go func() {
		send := func(event string, data any) error {
			if err := enc.Encode(data); err != nil {
				return fmt.Errorf("encode: %s", err)
			}
			return nil
		}

		for {
			select {
			case <-ctx.Done():
				return
			case rreq, ok := <-cout:
				if !ok {
					// closed channel
					return
				}
				mcpsrv.Handle(ctx, rreq, mcpconn, send)
			}
		}
	}()

	go func() {
		defer close(cin)
		bufin := bufio.NewReader(in)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				b, err := bufin.ReadBytes('\n')
				if err != nil {
					slog.Debug("stdin read", slog.Any("err", err))
					cancel()
					return
				}
				cin <- b
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			close(cout)
			return nil
		case b := <-cin:
			mcpreq, err := mcpsrv.Decode(bytes.NewReader(b))
			if err != nil {
				slog.Error("message decode error", slog.Any("err", err))
				// TODO return an error
				continue
			}

			cout <- mcpreq
		}
	}
}
