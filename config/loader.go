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
// rejecting unknown fields. Errors are prefixed "config:"; a yaml.TypeError is
// flattened into a single message joining each field-level error.
func Decode(r io.Reader, target any) error {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(target); err != nil {
		var typeErr *yaml.TypeError
		if errors.As(err, &typeErr) {
			return fmt.Errorf("config: %s", strings.Join(typeErr.Errors, "; "))
		}
		return fmt.Errorf("config: %w", err)
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
