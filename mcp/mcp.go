package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/lightpanda-io/go-mcp-demo/rpc"
)

// https://github.com/modelcontextprotocol/modelcontextprotocol/blob/main/schema/2025-03-26/schema.ts

const Version = "2024-11-05"

type Request any

func Decode(r rpc.Request) (Request, error) {
	switch r.Method {
	case InitializeMethod:
		rr := InitializeRequest{Request: r}
		if err := json.Unmarshal(r.Params, &rr.Params); err != nil {
			return nil, fmt.Errorf("decode: %w", err)
		}

		return rr, nil
	}

	return nil, fmt.Errorf("invalid mcp: %s", r.Method)
}

type Capability struct{}
type Capabilities map[string]Capability

const InitializeMethod = "initialize"

type Info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeRequest struct {
	rpc.Request
	Params struct {
		ProtocolVersion string       `json:"protocolVersion"`
		ClientInfo      Info         `json:"clientInfo"`
		Capabilities    Capabilities `json:"capabilities"`
	} `json:"params"`
}

type InitializeResponse struct {
	ProtocolVersion string       `json:"protocolVersion"`
	ServerInfo      Info         `json:"serverInfo"`
	Capabilities    Capabilities `json:"capabilities"`
}
