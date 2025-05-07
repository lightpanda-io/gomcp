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
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/chromedp/chromedp"
)

const (
	exitOK   = 0
	exitFail = 1
)

// main starts interruptable context and runs the program.
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	err := run(ctx, os.Args, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(exitFail)
	}

	os.Exit(exitOK)
}

const (
	ApiDefaultAddress = "127.0.0.1:8081"
	CdpWSDefault      = "ws://127.0.0.1:9222"
)

func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	// declare runtime flag parameters.
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.SetOutput(stderr)

	var (
		verbose = flags.Bool("verbose", false, "enable debug log level")
		apiaddr = flags.String("api-addr", env("MCP_API_ADDRESS", ApiDefaultAddress), "http api server address")
		cdpws   = flags.String("cdp", env("MCP_CDP", CdpWSDefault), "cdp ws to connect")
	)

	// usage func declaration.
	exec := args[0]
	flags.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s sse|stdio|download\n", exec)
		fmt.Fprintf(stderr, "Demo MCP server.\n")
		fmt.Fprintf(stderr, "\nCommands:\n")
		fmt.Fprintf(stderr, "\tstdio\t\tstarts the stdio server\n")
		fmt.Fprintf(stderr, "\tsse\t\tstarts the HTTP SSE MCP server\n")
		fmt.Fprintf(stderr, "\tdownload\tinstalls or updates the Lightpanda browser\n")
		fmt.Fprintf(stderr, "\nCommand line options:\n")
		flags.PrintDefaults()
		fmt.Fprintf(stderr, "\nEnvironment vars:\n")
		fmt.Fprintf(stderr, "\tMCP_API_ADDRESS\t\tdefault %s\n", ApiDefaultAddress)
		fmt.Fprintf(stderr, "\tMCP_CDP\t\t\tdefault %s\n", CdpWSDefault)
	}
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return errors.New("bad arguments")
	}

	if *verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	cdpctx, cancel := chromedp.NewRemoteAllocator(ctx,
		*cdpws, chromedp.NoModifyURL,
	)
	defer cancel()

	mcpsrv := NewMCPServer("lightpanda go mcp", "1.0.0", cdpctx)

	switch args[0] {
	case "download":
		// TODO download the lightpanda browser.
		return nil
	case "stdio":
		return runstd(ctx, stdin, stdout, mcpsrv)
	case "sse":
		return runapi(ctx, *apiaddr, mcpsrv)
	}

	flags.Usage()
	return errors.New("bad command")
}

// env returns the env value corresponding to the key or the default string.
func env(key, dflt string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return dflt
	}

	return val
}
