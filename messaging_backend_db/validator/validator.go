// Package validator is used to validate the documents against the given schema.
package validator

import (
	"bytes"
	"errors"
	"log/slog"

	"github.com/santhosh-tekuri/jsonschema"
)

// Validate validates the json file agains a json schema, returns true if document conforms to schema, else false.
func Validate(schema *jsonschema.Schema, docJSON []byte) (bool, error) {
	slog.Info("Validation: check that given schema is not nil")
	if schema == nil {
		slog.Error("schema is nil")
		return false, errors.New("given schema is not valid")
	}

	jsonReader := bytes.NewReader(docJSON)

	slog.Info("past NewReader docJSON")

	// Validate the JSON data against the compiled schema
	if err := schema.Validate(jsonReader); err != nil {
		slog.Error("data does not conform to the schema:", err)
		return false, err
	}

	slog.Info("data conforms to schema")
	return true, nil
}
