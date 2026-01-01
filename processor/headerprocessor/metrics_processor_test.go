// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package headerprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func TestMetricsProcessor(t *testing.T) {
	cfg := &Config{
		Headers: []HeaderConfig{
			{Name: "X-Test", Attribute: "test_attr"},
		},
	}
	he, err := newHeaderEnricher(cfg)
	require.NoError(t, err)

	mp := newMetricsProcessor(zap.NewNop(), he)

	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()

	// Add one of each metric type
	mGauge := sm.Metrics().AppendEmpty()
	mGauge.SetName("gauge")
	mGauge.SetEmptyGauge().DataPoints().AppendEmpty()

	mSum := sm.Metrics().AppendEmpty()
	mSum.SetName("sum")
	mSum.SetEmptySum().DataPoints().AppendEmpty()

	mHist := sm.Metrics().AppendEmpty()
	mHist.SetName("histogram")
	mHist.SetEmptyHistogram().DataPoints().AppendEmpty()

	mExpHist := sm.Metrics().AppendEmpty()
	mExpHist.SetName("exp_histogram")
	mExpHist.SetEmptyExponentialHistogram().DataPoints().AppendEmpty()

	mSummary := sm.Metrics().AppendEmpty()
	mSummary.SetName("summary")
	mSummary.SetEmptySummary().DataPoints().AppendEmpty()

	mdEmpty := sm.Metrics().AppendEmpty()
	mdEmpty.SetName("empty")

	ctx := client.NewContext(context.Background(), client.Info{
		Metadata: client.NewMetadata(map[string][]string{
			"X-Test": {"val1"},
		}),
	})

	processed, err := mp.processMetrics(ctx, md)
	assert.NoError(t, err)

	// Verify attributes are added to all data points
	rms := processed.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		sms := rms.At(i).ScopeMetrics()
		for j := 0; j < sms.Len(); j++ {
			ms := sms.At(j).Metrics()
			for k := 0; k < ms.Len(); k++ {
				m := ms.At(k)
				if m.Type() == pmetric.MetricTypeEmpty {
					continue
				}

				var dps pmetric.NumberDataPointSlice
				switch m.Type() {
				case pmetric.MetricTypeGauge:
					dps = m.Gauge().DataPoints()
				case pmetric.MetricTypeSum:
					dps = m.Sum().DataPoints()
				case pmetric.MetricTypeHistogram:
					hps := m.Histogram().DataPoints()
					for l := 0; l < hps.Len(); l++ {
						val, ok := hps.At(l).Attributes().Get("test_attr")
						assert.True(t, ok, "metric %s missing attribute", m.Name())
						assert.Equal(t, "val1", val.Str())
					}
					continue
				case pmetric.MetricTypeExponentialHistogram:
					ehps := m.ExponentialHistogram().DataPoints()
					for l := 0; l < ehps.Len(); l++ {
						val, ok := ehps.At(l).Attributes().Get("test_attr")
						assert.True(t, ok, "metric %s missing attribute", m.Name())
						assert.Equal(t, "val1", val.Str())
					}
					continue
				case pmetric.MetricTypeSummary:
					sps := m.Summary().DataPoints()
					for l := 0; l < sps.Len(); l++ {
						val, ok := sps.At(l).Attributes().Get("test_attr")
						assert.True(t, ok, "metric %s missing attribute", m.Name())
						assert.Equal(t, "val1", val.Str())
					}
					continue
				}

				for l := 0; l < dps.Len(); l++ {
					val, ok := dps.At(l).Attributes().Get("test_attr")
					assert.True(t, ok, "metric %s missing attribute", m.Name())
					assert.Equal(t, "val1", val.Str())
				}
			}
		}
	}
}

func TestMetricsProcessorNoHeaders(t *testing.T) {
	cfg := &Config{
		Headers: []HeaderConfig{
			{Name: "X-Test", Attribute: "test_attr"},
		},
	}
	he, err := newHeaderEnricher(cfg)
	require.NoError(t, err)

	mp := newMetricsProcessor(zap.NewNop(), he)

	md := pmetric.NewMetrics()
	m := md.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	m.SetEmptyGauge().DataPoints().AppendEmpty()

	// Context without headers
	ctx := context.Background()

	processed, err := mp.processMetrics(ctx, md)
	assert.NoError(t, err)

	// Verify no attributes added
	attrs := processed.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).Attributes()
	assert.Equal(t, 0, attrs.Len())
}
