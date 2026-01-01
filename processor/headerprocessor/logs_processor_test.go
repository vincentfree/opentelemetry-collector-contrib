// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package headerprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

func TestLogsProcessor(t *testing.T) {
	cfg := &Config{
		Headers: []HeaderConfig{
			{Name: "X-Test", Attribute: "test_attr"},
		},
	}
	he, err := newHeaderEnricher(cfg)
	require.NoError(t, err)

	lp := newLogsProcessor(zap.NewNop(), he)

	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	lr := sl.LogRecords().AppendEmpty()
	lr.Body().SetStr("test-log")

	ctx := client.NewContext(context.Background(), client.Info{
		Metadata: client.NewMetadata(map[string][]string{
			"X-Test": {"val1"},
		}),
	})

	processed, err := lp.processLogs(ctx, ld)
	assert.NoError(t, err)

	// Verify attributes are added to log records
	rls := processed.ResourceLogs()
	for i := 0; i < rls.Len(); i++ {
		sls := rls.At(i).ScopeLogs()
		for j := 0; j < sls.Len(); j++ {
			records := sls.At(j).LogRecords()
			for k := 0; k < records.Len(); k++ {
				attrs := records.At(k).Attributes()
				val, ok := attrs.Get("test_attr")
				assert.True(t, ok)
				assert.Equal(t, "val1", val.Str())
			}
		}
	}
}

func TestLogsProcessorNoHeaders(t *testing.T) {
	cfg := &Config{
		Headers: []HeaderConfig{
			{Name: "X-Test", Attribute: "test_attr"},
		},
	}
	he, err := newHeaderEnricher(cfg)
	require.NoError(t, err)

	lp := newLogsProcessor(zap.NewNop(), he)

	ld := plog.NewLogs()
	ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()

	// Context without headers
	ctx := context.Background()

	processed, err := lp.processLogs(ctx, ld)
	assert.NoError(t, err)

	// Verify no attributes added
	attrs := processed.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes()
	assert.Equal(t, 0, attrs.Len())
}
