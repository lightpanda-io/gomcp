package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lightpanda-io/go-mcp-demo/mcp"
)

type MCPTool interface {
	Name() string
	Description() string
	Properties() mcp.Properties
	Call(json.RawMessage) (string, error)
}

type MCPServer struct {
	Name    string
	Version string

	Tools []MCPTool
}

func (s *MCPServer) ListTools() []mcp.Tool {
	var tools []mcp.Tool
	for _, t := range s.Tools {
		tool := mcp.Tool{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: mcp.NewSchemaObject(t.Properties()),
		}
		tools = append(tools, tool)
	}

	return tools
}

var ErrNoTool = errors.New("no tool found")

func (s *MCPServer) CallTool(req mcp.ToolsCallRequest) (string, error) {
	for _, t := range s.Tools {
		if t.Name() == req.Params.Name {
			res, err := t.Call(req.Params.Arguments)
			if err != nil {
				return "", fmt.Errorf("call %s: %w", t.Name(), err)
			}

			return res, nil
		}
	}

	// no tool found
	return "", ErrNoTool
}
