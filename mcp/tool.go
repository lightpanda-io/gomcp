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

type Schema any

type SchemaType struct {
	Type string `json:"type"`
}

type schemaString SchemaType

func NewSchemaString() schemaString {
	return schemaString(SchemaType{Type: "string"})
}

type Properties map[string]Schema

type schemaObject struct {
	SchemaType
	Properties Properties `json:"properties"`
}

func NewSchemaObject(p map[string]Schema) schemaObject {
	return schemaObject{
		SchemaType: SchemaType{Type: "object"},
		Properties: p,
	}
}

type Tool struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	InputSchema schemaObject `json:"inputSchema"`
	// TODO annotations
}
