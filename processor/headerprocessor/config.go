// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package headerprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/headerprocessor"

import (
	"errors"
)

// Config defines the configuration for the header processor.
// It controls which headers to extract from client context and how to map
// them to attribute keys on telemetry data.
type Config struct {
	// Headers is a list of header extraction rules.
	Headers []HeaderConfig `mapstructure:"headers"`

	// GlobalPrefix is applied to all headers unless a per-header Prefix is provided.
	GlobalPrefix string `mapstructure:"prefix"`

	// Separator used to join multiple header values. Defaults to ";" if empty.
	Separator string `mapstructure:"separator"`

	// IncludeAll extracts every header present in the client metadata.
	IncludeAll bool `mapstructure:"include_all"`

	// ExcludePatterns contains regex patterns to filter headers when IncludeAll is true.
	ExcludePatterns []string `mapstructure:"exclude_patterns"`

	// prevent unkeyed literal initialization
	_ struct{}
}

// HeaderConfig configures extraction for a single header.
type HeaderConfig struct {
	// Name of the HTTP header to extract (case-insensitive).
	Name string `mapstructure:"name"`
	// Attribute is the attribute name to set. If empty, defaults to the header Name.
	Attribute string `mapstructure:"attribute"`
	// Prefix to apply for this specific header. If empty, GlobalPrefix is used.
	Prefix string `mapstructure:"prefix"`

	// prevent unkeyed literal initialization
	_ struct{}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if !c.IncludeAll && len(c.Headers) == 0 {
		return errors.New("missing required field \"headers\" or set include_all: true")
	}
	for _, hc := range c.Headers {
		if hc.Name == "" {
			return errors.New("header entry missing required field \"name\"")
		}
	}
	return nil
}
