// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package headerprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

func TestTracesProcessor(t *testing.T) {
	cfg := &Config{
		Headers: []HeaderConfig{
			{Name: "X-Test", Attribute: "test_attr"},
		},
	}
	he, err := newHeaderEnricher(cfg)
	require.NoError(t, err)

	tp := newTracesProcessor(zap.NewNop(), he)

	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	ss := rs.ScopeSpans().AppendEmpty()
	span := ss.Spans().AppendEmpty()
	span.SetName("test-span")

	ctx := client.NewContext(context.Background(), client.Info{
		Metadata: client.NewMetadata(map[string][]string{
			"X-Test": {"val1"},
		}),
	})

	processed, err := tp.processTraces(ctx, td)
	assert.NoError(t, err)

	// Verify attributes are added to spans
	rss := processed.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		sss := rss.At(i).ScopeSpans()
		for j := 0; j < sss.Len(); j++ {
			spans := sss.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				attrs := spans.At(k).Attributes()
				val, ok := attrs.Get("test_attr")
				assert.True(t, ok)
				assert.Equal(t, "val1", val.Str())
			}
		}
	}
}

func TestTracesProcessorNoHeaders(t *testing.T) {
	cfg := &Config{
		Headers: []HeaderConfig{
			{Name: "X-Test", Attribute: "test_attr"},
		},
	}
	he, err := newHeaderEnricher(cfg)
	require.NoError(t, err)

	tp := newTracesProcessor(zap.NewNop(), he)

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()

	// Context without headers
	ctx := context.Background()

	processed, err := tp.processTraces(ctx, td)
	assert.NoError(t, err)

	// Verify no attributes added
	attrs := processed.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes()
	assert.Equal(t, 0, attrs.Len())
}
