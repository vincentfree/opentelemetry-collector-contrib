// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package headerprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor/processortest"
)

func TestIntegrationTraces(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.Headers = []HeaderConfig{
		{Name: "X-Test-Header", Attribute: "test.header"},
	}

	next := new(consumertest.TracesSink)
	set := processortest.NewNopSettings(factory.Type())
	tp, err := factory.CreateTraces(context.Background(), set, cfg, next)
	require.NoError(t, err)

	td := ptrace.NewTraces()
	span := td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.SetName("test-span")

	md := client.NewMetadata(map[string][]string{
		"X-Test-Header": {"test-value"},
	})
	ctx := client.NewContext(context.Background(), client.Info{Metadata: md})

	err = tp.ConsumeTraces(ctx, td)
	require.NoError(t, err)

	got := next.AllTraces()
	require.Len(t, got, 1)
	resAttrs := got[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes()
	val, ok := resAttrs.Get("test.header")
	assert.True(t, ok)
	assert.Equal(t, "test-value", val.Str())
}

func TestIntegrationMetrics(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.IncludeAll = true
	cfg.GlobalPrefix = "header."

	next := new(consumertest.MetricsSink)
	set := processortest.NewNopSettings(factory.Type())
	mp, err := factory.CreateMetrics(context.Background(), set, cfg, next)
	require.NoError(t, err)

	md := pmetric.NewMetrics()
	gauge := md.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	gauge.SetName("test-metric")
	gauge.SetEmptyGauge().DataPoints().AppendEmpty().SetDoubleValue(1.0)

	metadata := client.NewMetadata(map[string][]string{
		"X-Custom": {"val1", "val2"},
	})
	ctx := client.NewContext(context.Background(), client.Info{Metadata: metadata})

	err = mp.ConsumeMetrics(ctx, md)
	require.NoError(t, err)

	got := next.AllMetrics()
	require.Len(t, got, 1)
	attrs := got[0].ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).Attributes()
	val, ok := attrs.Get("header.x-custom")
	assert.True(t, ok)
	assert.Equal(t, "val1;val2", val.Str())
}

func TestIntegrationLogs(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.Headers = []HeaderConfig{
		{Name: "X-Log-Header", Attribute: "log.attr", Prefix: "custom."},
	}

	next := new(consumertest.LogsSink)
	set := processortest.NewNopSettings(factory.Type())
	lp, err := factory.CreateLogs(context.Background(), set, cfg, next)
	require.NoError(t, err)

	ld := plog.NewLogs()
	log := ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	log.Body().SetStr("test-log")

	md := client.NewMetadata(map[string][]string{
		"X-Log-Header": {"log-value"},
	})
	ctx := client.NewContext(context.Background(), client.Info{Metadata: md})

	err = lp.ConsumeLogs(ctx, ld)
	require.NoError(t, err)

	got := next.AllLogs()
	require.Len(t, got, 1)
	attrs := got[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes()
	val, ok := attrs.Get("custom.log.attr")
	assert.True(t, ok)
	assert.Equal(t, "log-value", val.Str())
}

func TestIntegrationIncludeAllWithExclusion(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.IncludeAll = true
	cfg.ExcludePatterns = []string{"^auth.*", "sensitive"}

	next := new(consumertest.TracesSink)
	set := processortest.NewNopSettings(factory.Type())
	tp, err := factory.CreateTraces(context.Background(), set, cfg, next)
	require.NoError(t, err)

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()

	md := client.NewMetadata(map[string][]string{
		"Authorization": {"secret"},
		"Sensitive":     {"data"},
		"Public":        {"visible"},
	})
	ctx := client.NewContext(context.Background(), client.Info{Metadata: md})

	err = tp.ConsumeTraces(ctx, td)
	require.NoError(t, err)

	got := next.AllTraces()
	require.Len(t, got, 1)
	attrs := got[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes()

	_, ok := attrs.Get("authorization")
	assert.False(t, ok)
	_, ok = attrs.Get("sensitive")
	assert.False(t, ok)
	val, ok := attrs.Get("public")
	assert.True(t, ok)
	assert.Equal(t, "visible", val.Str())
}
