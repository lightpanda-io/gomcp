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
