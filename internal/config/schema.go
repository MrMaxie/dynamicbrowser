package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed config.schema.json
var schema []byte

var configurationSchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema))
	if err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	const location = "https://dynamicbrowser.invalid/config.schema.json"
	if err := compiler.AddResource(location, document); err != nil {
		return nil, err
	}
	return compiler.Compile(location)
})

func validateConfiguration(values map[string]any) error {
	compiled, err := configurationSchema()
	if err != nil {
		return err
	}
	return compiled.Validate(values)
}

const initialConfiguration = "# yaml-language-server: $schema=./config.schema.json\n{}\n"

func ensureFiles(path string) error {
	if err := createIfMissing(filepath.Join(filepath.Dir(path), "config.schema.json"), schema); err != nil {
		return fmt.Errorf("create config.schema.json: %w", err)
	}
	if err := createIfMissing(path, []byte(initialConfiguration)); err != nil {
		return fmt.Errorf("create config.yaml: %w", err)
	}
	return nil
}

func createIfMissing(path string, content []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = file.Write(content)
	return errors.Join(err, file.Close())
}
