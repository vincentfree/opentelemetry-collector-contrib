// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package headerprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/headerprocessor"

import (
	"context"
	"regexp"
	"strings"

	"go.opentelemetry.io/collector/client"
)

type headerEnricher struct {
	cfg             *Config
	sep             string
	excludeMatchers []*regexp.Regexp
}

func newHeaderEnricher(cfg *Config) (*headerEnricher, error) {
	sep := cfg.Separator
	if sep == "" {
		sep = ";"
	}

	patterns := cfg.ExcludePatterns
	if cfg.IncludeAll && len(patterns) == 0 {
		// default sensitive headers to exclude
		patterns = []string{"^authorization$", "^cookie$"}
	}

	matchers := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		r, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, r)
	}

	return &headerEnricher{cfg: cfg, sep: sep, excludeMatchers: matchers}, nil
}

// attributesFromContext extracts headers from client metadata in the context and
// returns a map of attribute name -> value according to the configuration.
func (he *headerEnricher) attributesFromContext(ctx context.Context) map[string]string {
	info := client.FromContext(ctx)
	md := info.Metadata

	attrs := map[string]string{}

	if he.cfg.IncludeAll {
		// Iterate all keys
		for k := range md.Keys() {
			if he.excluded(strings.ToLower(k)) {
				continue
			}
			values := md.Get(k)
			if len(values) == 0 {
				continue
			}
			key := he.buildAttrKey("", k)
			attrs[key] = strings.Join(values, he.sep)
		}
		if len(attrs) == 0 {
			return nil
		}
		return attrs
	}

	// Extract only configured headers (case-insensitive)
	// Build a map of lowercase header name to actual key in metadata for quick lookup
	lowerToActual := map[string]string{}
	for k := range md.Keys() {
		lowerToActual[strings.ToLower(k)] = k
	}
	for _, hc := range he.cfg.Headers {
		lk := strings.ToLower(hc.Name)
		actual, ok := lowerToActual[lk]
		if !ok {
			continue
		}
		values := md.Get(actual)
		if len(values) == 0 {
			continue
		}
		key := he.buildAttrKey(hc.Prefix, firstNonEmpty(hc.Attribute, hc.Name))
		attrs[key] = strings.Join(values, he.sep)
	}
	if len(attrs) == 0 {
		return nil
	}
	return attrs
}

func (he *headerEnricher) excluded(lowerHeader string) bool {
	for _, r := range he.excludeMatchers {
		if r.MatchString(lowerHeader) {
			return true
		}
	}
	return false
}

func (he *headerEnricher) buildAttrKey(perHeaderPrefix, name string) string {
	prefix := strings.TrimSpace(perHeaderPrefix)
	if prefix == "" {
		prefix = strings.TrimSpace(he.cfg.GlobalPrefix)
	}
	return prefix + name
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
