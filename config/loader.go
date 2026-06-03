// Package config provides shared configuration primitives for the adaouat CLIs:
// a strict YAML loader, app-name-parameterized path resolution, and structured
// validation errors. Schemas, defaults, and merge semantics stay in the apps.
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Decode strictly parses YAML from r into target (a pointer to a struct),
// rejecting unknown fields. A yaml.TypeError is flattened into a single error
// joining each field-level message; other errors are returned as-is. Callers
// add their own context prefix.
func Decode(r io.Reader, target any) error {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(target); err != nil {
		var typeErr *yaml.TypeError
		if errors.As(err, &typeErr) {
			return errors.New(strings.Join(typeErr.Errors, "; "))
		}
		return err
	}
	return nil
}

// Load opens path and decodes it into target via Decode.
func Load(path string, target any) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	return Decode(f, target)
}
