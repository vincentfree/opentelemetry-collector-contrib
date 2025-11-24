// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package headerprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/headerprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type metricsProcessor struct {
	logger   *zap.Logger
	enricher *headerEnricher
}

func newMetricsProcessor(logger *zap.Logger, he *headerEnricher) *metricsProcessor {
	return &metricsProcessor{logger: logger, enricher: he}
}

func (p *metricsProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	kvs := p.enricher.attributesFromContext(ctx)
	if len(kvs) == 0 {
		return md, nil
	}
	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rs := rms.At(i)
		ilms := rs.ScopeMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ils := ilms.At(j)
			metrics := ils.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				m := metrics.At(k)
				p.applyToMetric(m, kvs)
			}
		}
	}
	return md, nil
}

func (p *metricsProcessor) applyToMetric(m pmetric.Metric, kvs map[string]string) {
	switch m.Type() { //exhaustive:enforce
	case pmetric.MetricTypeGauge:
		dps := m.Gauge().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			attrs := dps.At(i).Attributes()
			for k, v := range kvs {
				attrs.PutStr(k, v)
			}
		}
	case pmetric.MetricTypeSum:
		dps := m.Sum().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			attrs := dps.At(i).Attributes()
			for k, v := range kvs {
				attrs.PutStr(k, v)
			}
		}
	case pmetric.MetricTypeHistogram:
		dps := m.Histogram().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			attrs := dps.At(i).Attributes()
			for k, v := range kvs {
				attrs.PutStr(k, v)
			}
		}
	case pmetric.MetricTypeExponentialHistogram:
		dps := m.ExponentialHistogram().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			attrs := dps.At(i).Attributes()
			for k, v := range kvs {
				attrs.PutStr(k, v)
			}
		}
	case pmetric.MetricTypeSummary:
		dps := m.Summary().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			attrs := dps.At(i).Attributes()
			for k, v := range kvs {
				attrs.PutStr(k, v)
			}
		}
	case pmetric.MetricTypeEmpty:
	}
}
