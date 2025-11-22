// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package baggageprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/baggageprocessor"

import (
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/component"
)

// ActionType defines the type of action to perform on baggage entries
type ActionType string

const (
	// Extract extracts baggage entries from the context and adds them as attributes
	Extract ActionType = "extract"
	// Inject injects attributes as baggage entries into the context
	Inject ActionType = "inject"
	// Update updates existing baggage entries with new values
	Update ActionType = "update"
	// Upsert inserts or updates baggage entries
	Upsert ActionType = "upsert"
	// Delete removes baggage entries from the context
	Delete ActionType = "delete"
)

// Action defines a single baggage processing action
type Action struct {
	// Key is the baggage key to operate on
	Key string `mapstructure:"key"`

	// Action specifies the type of action to perform
	Action ActionType `mapstructure:"action"`

	// Value is the value to set for inject/update/upsert actions
	Value string `mapstructure:"value,omitempty"`

	// FromAttribute specifies the attribute name to get the value from for inject actions
	FromAttribute string `mapstructure:"from_attribute,omitempty"`

	// FromContext indicates whether to extract from baggage context (for extract action)
	FromContext bool `mapstructure:"from_context,omitempty"`

	// ToAttribute specifies the attribute name to set the value to (for extract action)
	ToAttribute string `mapstructure:"to_attribute,omitempty"`

	// Properties defines W3C baggage properties to set (key=value pairs)
	Properties map[string]string `mapstructure:"properties,omitempty"`
}

// Config defines the configuration for the baggage processor
type Config struct {
	// Actions is the list of actions to perform on baggage entries
	Actions []Action `mapstructure:"actions"`

	// AttributePrefix is the prefix to add to attribute names when extracting baggage
	// Default is "baggage."
	AttributePrefix string `mapstructure:"attribute_prefix,omitempty"`

	// MaxBaggageSize is the maximum size in bytes for the baggage header
	// Default is 8192 bytes as per W3C specification
	MaxBaggageSize int `mapstructure:"max_baggage_size,omitempty"`

	// DropInvalidBaggage indicates whether to drop invalid baggage entries
	// instead of failing the entire operation
	DropInvalidBaggage bool `mapstructure:"drop_invalid_baggage,omitempty"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if len(cfg.Actions) == 0 {
		return errors.New("at least one action must be specified")
	}

	for i, action := range cfg.Actions {
		if err := action.validate(); err != nil {
			return fmt.Errorf("action %d: %w", i, err)
		}
	}

	if cfg.MaxBaggageSize < 0 {
		return errors.New("max_baggage_size must be non-negative")
	}

	return nil
}

// validate checks if an action configuration is valid
func (a *Action) validate() error {
	if a.Key == "" {
		return errors.New("key must be specified")
	}

	switch a.Action {
	case Extract:
		if !a.FromContext {
			return errors.New("from_context must be true for extract action")
		}
	case Inject:
		if a.Value == "" && a.FromAttribute == "" {
			return errors.New("either value or from_attribute must be specified for inject action")
		}
		if a.Value != "" && a.FromAttribute != "" {
			return errors.New("only one of value or from_attribute can be specified for inject action")
		}
	case Update, Upsert:
		if a.Value == "" && a.FromAttribute == "" {
			return errors.New("either value or from_attribute must be specified for update/upsert action")
		}
		if a.Value != "" && a.FromAttribute != "" {
			return errors.New("only one of value or from_attribute can be specified for update/upsert action")
		}
	case Delete:
		// No additional validation needed for delete
	default:
		return fmt.Errorf("unknown action type: %s", a.Action)
	}

	return nil
}
