// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package headerprocessor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid specific headers",
			cfg: &Config{
				Headers: []HeaderConfig{
					{Name: "X-Test", Attribute: "test"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid include_all",
			cfg: &Config{
				IncludeAll: true,
			},
			wantErr: false,
		},
		{
			name: "missing headers and include_all false",
			cfg: &Config{
				IncludeAll: false,
				Headers:    nil,
			},
			wantErr: true,
		},
		{
			name: "header missing name",
			cfg: &Config{
				Headers: []HeaderConfig{
					{Attribute: "test"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid multiple headers",
			cfg: &Config{
				Headers: []HeaderConfig{
					{Name: "X-Test-1"},
					{Name: "X-Test-2"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
