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
