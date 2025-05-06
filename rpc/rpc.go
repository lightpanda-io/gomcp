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

package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
)

const Version = "2.0"

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Request struct {
	Version string          `json:"jsonrpc"`
	Id      int             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params"`
	Error   *Error          `json:"error,omitempty"`
}

var InvalidRequestErr = errors.New("invalid request")

func (req Request) Validate() error {
	if req.Version != Version {
		return InvalidRequestErr
	}

	if req.Method == "" && req.Error == nil {
		return InvalidRequestErr
	}

	return nil
}

func (req Request) Err() error {
	if req.Error == nil {
		return nil
	}

	return fmt.Errorf("code %d: %s", req.Error.Code, req.Error.Message)
}

type Response struct {
	Version string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  any    `json:"result"`
}

func NewResponse(data any, id int) Response {
	return Response{
		Result:  data,
		Id:      id,
		Version: Version,
	}
}
