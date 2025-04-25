package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/chromedp/chromedp"
	"github.com/lightpanda-io/gomcp/mcp"
)

// A connection with a client
type MCPConn struct {
	srv       *MCPServer
	cdpctx    context.Context
	cdpcancel context.CancelFunc
}

func (c *MCPConn) Close() {
	if c.cdpcancel != nil {
		c.cdpcancel()
	}
}

func (c *MCPConn) connect() error {
	if c.cdpcancel != nil {
		c.cdpcancel()
	}

	ctx, cancel := chromedp.NewContext(c.srv.cdpctx)

	// ensure the first tab is created
	if err := chromedp.Run(ctx); err != nil {
		cancel()
		return fmt.Errorf("new tab: %w", err)
	}

	c.cdpctx = ctx
	c.cdpcancel = cancel

	return nil
}

// Navigate to a specified URL
func (c *MCPConn) Goto(url string) (string, error) {

	if err := c.connect(); err != nil {
		return "", fmt.Errorf("browser connect: %w", err)
	}

	err := chromedp.Run(c.cdpctx, chromedp.Navigate(url))
	if err != nil {
		return "", fmt.Errorf("navigate %s: %w", url, err)
	}

	return fmt.Sprintf("navigation to %s done", url), nil
}

// Return the document's HTML
func (c *MCPConn) GetHTML() (string, error) {
	if c.cdpctx == nil {
		return "", errors.New("no browser connection, try to use goto first")
	}

	var content string
	err := chromedp.Run(c.cdpctx, chromedp.OuterHTML("html", &content))
	if err != nil {
		return "", fmt.Errorf("outerHTML: %w", err)
	}

	return content, nil
}

type MCPServer struct {
	Name    string
	Version string

	cdpctx context.Context
}

func NewMCPServer(name, version string, cdpctx context.Context) *MCPServer {
	return &MCPServer{
		Name:    name,
		Version: version,
		cdpctx:  cdpctx,
	}
}

func (s *MCPServer) NewConn() *MCPConn {
	return &MCPConn{
		srv: s,
	}
}

func (s *MCPServer) ListTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "goto",
			Description: "Navigate to a specified URL and load the page in" +
				"memory so it can be reused later for info extraction",
			InputSchema: mcp.NewSchemaObject(mcp.Properties{
				"url": mcp.NewSchemaString(),
			}),
		},
		{
			Name:        "html",
			Description: "Get the full HTML of the opened page",
			InputSchema: mcp.NewSchemaObject(mcp.Properties{}),
		},
	}
}

var ErrNoTool = errors.New("no tool found")

func (s *MCPServer) CallTool(ctx context.Context, conn *MCPConn, req mcp.ToolsCallRequest) (string, error) {
	v := req.Params.Arguments

	switch req.Params.Name {
	case "goto":
		var args struct {
			URL string `json:"url"`
		}

		if err := json.Unmarshal(v, &args); err != nil {
			return "", fmt.Errorf("args decode: %w", err)
		}

		if args.URL == "" {
			return "", errors.New("no url")
		}
		return conn.Goto(args.URL)
	case "html":
		return conn.GetHTML()
	}

	// no tool found
	return "", ErrNoTool
}
