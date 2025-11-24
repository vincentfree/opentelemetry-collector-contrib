package headerprocessor

import (
	"context"
	"testing"

	"go.opentelemetry.io/collector/client"
)

func TestEnricherSpecificHeaders(t *testing.T) {
	cfg := &Config{
		Headers: []HeaderConfig{
			{Name: "User-Agent", Attribute: "http.user_agent"},
			{Name: "x-correlation-id", Prefix: "req."},
		},
		GlobalPrefix: "",
		Separator:    ";",
	}
	he, err := newHeaderEnricher(cfg)
	if err != nil {
		t.Fatalf("newHeaderEnricher error: %v", err)
	}

	md := client.NewMetadata(map[string][]string{
		"user-agent":       {"UA-1"},
		"X-Correlation-Id": {"abc", "def"},
		"ignore":           {"nope"},
	})
	ctx := client.NewContext(context.Background(), client.Info{Metadata: md})

	attrs := he.attributesFromContext(ctx)

	if got := attrs["http.user_agent"]; got != "UA-1" {
		t.Fatalf("user agent attr mismatch: %q", got)
	}
	if got := attrs["req.x-correlation-id"]; got != "abc;def" {
		t.Fatalf("correlation attr mismatch: %q", got)
	}
	if _, ok := attrs["ignore"]; ok {
		t.Fatalf("unexpected attr for ignored header")
	}
}

func TestEnricherIncludeAllWithExclude(t *testing.T) {
	cfg := &Config{
		IncludeAll:      true,
		GlobalPrefix:    "http.header.",
		Separator:       ";",
		ExcludePatterns: []string{"^authorization$", "^x-forwarded-.*$"},
	}
	he, err := newHeaderEnricher(cfg)
	if err != nil {
		t.Fatalf("newHeaderEnricher error: %v", err)
	}
	md := client.NewMetadata(map[string][]string{
		"Authorization":   {"secret"},
		"User-Agent":      {"UA-1"},
		"x-forwarded-for": {"10.0.0.1"},
	})
	ctx := client.NewContext(context.Background(), client.Info{Metadata: md})

	attrs := he.attributesFromContext(ctx)

	if _, ok := attrs["http.header.authorization"]; ok {
		t.Fatalf("authorization should be excluded")
	}
	if _, ok := attrs["http.header.x-forwarded-for"]; ok {
		t.Fatalf("x-forwarded-for should be excluded")
	}
	if got := attrs["http.header.user-agent"]; got != "UA-1" {
		t.Fatalf("expected user-agent attr, got %q", got)
	}
}
