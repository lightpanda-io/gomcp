package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"

	"github.com/lightpanda-io/gomcp/mcp"
	"github.com/lightpanda-io/gomcp/rpc"
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

// Return all links from a page
func (c *MCPConn) GetLinks() ([]string, error) {
	if c.cdpctx == nil {
		return nil, errors.New("no browser connection, try to use goto first")
	}

	var a []*cdp.Node
	if err := chromedp.Run(c.cdpctx, chromedp.Nodes(`a[href]`, &a)); err != nil {
		return nil, fmt.Errorf("get links: %w", err)
	}

	links := make([]string, 0, len(a))
	for _, aa := range a {
		v, ok := aa.Attribute("href")
		if ok {
			links = append(links, v)
		}
	}

	return links, nil
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
		{
			Name:        "links",
			Description: "list all links visibles in the opened page",
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
	case "links":
		links, err := conn.GetLinks()
		if err != nil {
			return "", err
		}
		return strings.Join(links, "\n"), nil
	}

	// no tool found
	return "", ErrNoTool
}

var ErrRPCRequest = errors.New("rpc request error")

// Decode a message
func (s *MCPServer) Decode(in io.Reader) (mcp.Request, error) {
	var empty mcp.Request

	dec := json.NewDecoder(in)
	var rreq rpc.Request
	if err := dec.Decode(&rreq); err != nil {
		return empty, fmt.Errorf("json decode: %w", err)
	}

	if err := rreq.Validate(); err != nil {
		return empty, fmt.Errorf("rpc validate: %w", err)
	}

	// The rpc request contains an error.
	if err := rreq.Err(); err != nil {
		return empty, errors.Join(ErrRPCRequest, rreq.Err())
	}

	mcpreq, err := mcp.Decode(rreq)
	if err != nil {
		return empty, fmt.Errorf("mcp validate: %w", err)
	}

	return mcpreq, err
}

type SendFn func(string, any) error

func (s *MCPServer) Handle(
	ctx context.Context,
	rreq mcp.Request,
	mcpconn *MCPConn,
	send SendFn,
) error {
	var senderr error
	switch r := rreq.(type) {
	case mcp.InitializeRequest:
		senderr = send("message", rpc.NewResponse(mcp.InitializeResponse{
			ProtocolVersion: mcp.Version,
			ServerInfo: mcp.Info{
				Name:    "lightpanda go mcp",
				Version: "1.0.0",
			},
			Capabilities: mcp.Capabilities{"tools": mcp.Capability{}},
		}, r.Request.Id))
	case mcp.ToolsListRequest:
		senderr = send("message", rpc.NewResponse(mcp.ToolsListResponse{
			Tools: s.ListTools(),
		}, r.Id))
	case mcp.ToolsCallRequest:
		slog.Debug("call tool", slog.String("name", r.Params.Name), slog.Int("id", r.Id))
		go func() {
			res, err := s.CallTool(ctx, mcpconn, r)

			if err != nil {
				slog.Error("call tool", slog.String("name", r.Params.Name), slog.Any("err", err))
				senderr = send("message", rpc.NewResponse(mcp.ToolsCallResponse{
					IsError: true,
					Content: []mcp.ToolsCallContent{{
						Type: "text",
						Text: err.Error(),
					}},
				}, r.Id))
			}

			senderr = send("message", rpc.NewResponse(mcp.ToolsCallResponse{
				Content: []mcp.ToolsCallContent{{
					Type: "text",
					Text: res,
				}},
			}, r.Id))
		}()

	case mcp.NotificationsCancelledRequest:
		slog.Debug("cancelled",
			slog.Int("id", r.Params.RequestId),
			slog.String("reason", r.Params.Reason),
		)
		// TODO cancel the corresponding request.
	}

	if senderr != nil {
		return fmt.Errorf("send message: %w", senderr)
	}

	return nil
}
