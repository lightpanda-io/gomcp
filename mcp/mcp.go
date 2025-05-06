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

package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/lightpanda-io/gomcp/rpc"
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
	case NotificationsInitializedMethod:
		return NotificationsInitializedRequest(r), nil
	case NotificationsCancelledMethod:
		rr := NotificationsCancelledRequest{Request: r}
		if err := json.Unmarshal(r.Params, &rr.Params); err != nil {
			return nil, fmt.Errorf("decode: %w", err)
		}

		return rr, nil
	case ToolsListMethod:
		return ToolsListRequest(r), nil
	case ToolsCallMethod:
		rr := ToolsCallRequest{Request: r}
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

const NotificationsInitializedMethod = "notifications/initialized"

type NotificationsInitializedRequest rpc.Request

const NotificationsCancelledMethod = "notifications/cancelled"

type NotificationsCancelledRequest struct {
	rpc.Request
	Params struct {
		RequestId int    `json:"requestId"`
		Reason    string `json:"reason"`
	}
}

const ToolsListMethod = "tools/list"

type ToolsListRequest rpc.Request

type ToolsListResponse struct {
	Tools []Tool `json:"tools"`
}

const ToolsCallMethod = "tools/call"

type ToolsCallRequest struct {
	rpc.Request
	Params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
		Meta      struct {
			ProgressToken int `json:"progressToken"`
		} `json:"_meta"`
	} `json:"params"`
}

type ToolsCallContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ToolsCallResponse struct {
	IsError bool               `json:"isError"`
	Content []ToolsCallContent `json:"content"`
}
