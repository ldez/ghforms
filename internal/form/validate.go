package form

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

//go:embed schema-issue-form.json
var jsonSchemaIssueForm []byte

//go:embed schema-discussion-form.json
var jsonSchemaDiscussionForm []byte

//go:embed schema-config.json
var jsonSchemaChooser []byte

type SchemaValidator struct {
	schema *jsonschema.Schema
}

func NewSchemaIssueValidator() (*SchemaValidator, error) {
	schema, err := getSchema(jsonSchemaIssueForm, "schema-issue-form.json")
	if err != nil {
		return nil, err
	}

	return &SchemaValidator{schema: schema}, nil
}

func NewSchemaDiscussionValidator() (*SchemaValidator, error) {
	schema, err := getSchema(jsonSchemaDiscussionForm, "schema-discussion-form.json")
	if err != nil {
		return nil, err
	}

	return &SchemaValidator{schema: schema}, nil
}

func NewSchemaChooserValidator() (*SchemaValidator, error) {
	schema, err := getSchema(jsonSchemaChooser, "schema-config.json")
	if err != nil {
		return nil, err
	}

	return &SchemaValidator{schema: schema}, nil
}

func (v *SchemaValidator) validate(source string, data []byte) error {
	var doc any

	err := yaml.Unmarshal(data, &doc)
	if err != nil {
		return fmt.Errorf("%s: invalid YAML: %w", source, err)
	}

	err = v.schema.Validate(doc)
	if err == nil {
		return nil
	}

	if ve, ok := errors.AsType[*jsonschema.ValidationError](err); ok {
		return fmt.Errorf("%s: schema validation failed:\n%s", source, strings.TrimRight(ve.Error(), "\n"))
	}

	return fmt.Errorf("%s: schema validation failed: %w", source, err)
}

func getSchema(data []byte, name string) (*jsonschema.Schema, error) {
	var raw any

	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, fmt.Errorf("invalid embedded schema: %w", err)
	}

	compiler := jsonschema.NewCompiler()

	err = compiler.AddResource(name, raw)
	if err != nil {
		return nil, fmt.Errorf("registering schema: %w", err)
	}

	schema, err := compiler.Compile(name)
	if err != nil {
		return nil, fmt.Errorf("compiling schema: %w", err)
	}

	return schema, nil
}
