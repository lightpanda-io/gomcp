package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lightpanda-io/go-mcp-demo/mcp"
)

type Helloworld struct{}

func (h Helloworld) Name() string {
	return "Hello World"
}
func (h Helloworld) Description() string {
	return "Say hello to the world"
}
func (h Helloworld) Properties() mcp.Properties {
	return mcp.Properties{
		"name": mcp.NewSchemaString(),
	}
}
func (h Helloworld) Call(v json.RawMessage) (string, error) {

	var args struct {
		Name string
	}

	if err := json.Unmarshal(v, &args); err != nil {
		return "", fmt.Errorf("args decode: %w", err)
	}

	if args.Name == "" {
		return "", errors.New("invalid arg")
	}

	return "Hello " + args.Name, nil
}
