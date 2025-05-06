// Copyright 2025 Lightpanda (Selecy SAS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
				if err := mcpsrv.Handle(ctx, rreq, mcpconn, send); err != nil {
					// disconnect on error
					slog.Error("handle req", slog.Any("err", err))
					return
				}
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
