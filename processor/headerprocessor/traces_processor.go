// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package headerprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/headerprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type tracesProcessor struct {
	logger   *zap.Logger
	enricher *headerEnricher
}

func newTracesProcessor(logger *zap.Logger, he *headerEnricher) *tracesProcessor {
	return &tracesProcessor{logger: logger, enricher: he}
}

func (p *tracesProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	kvs := p.enricher.attributesFromContext(ctx)
	if len(kvs) == 0 {
		return td, nil
	}
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		ilss := rs.ScopeSpans()
		for j := 0; j < ilss.Len(); j++ {
			ils := ilss.At(j)
			spans := ils.Spans()
			for k := 0; k < spans.Len(); k++ {
				attrs := spans.At(k).Attributes()
				for kName, v := range kvs {
					attrs.PutStr(kName, v)
				}
			}
		}
	}
	return td, nil
}
