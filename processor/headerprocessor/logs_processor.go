// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package headerprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/headerprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type logsProcessor struct {
	logger   *zap.Logger
	enricher *headerEnricher
}

func newLogsProcessor(logger *zap.Logger, he *headerEnricher) *logsProcessor {
	return &logsProcessor{logger: logger, enricher: he}
}

func (p *logsProcessor) processLogs(ctx context.Context, ld plog.Logs) (plog.Logs, error) {
	kvs := p.enricher.attributesFromContext(ctx)
	if len(kvs) == 0 {
		return ld, nil
	}
	rlss := ld.ResourceLogs()
	for i := 0; i < rlss.Len(); i++ {
		rls := rlss.At(i)
		slss := rls.ScopeLogs()
		for j := 0; j < slss.Len(); j++ {
			sls := slss.At(j)
			records := sls.LogRecords()
			for k := 0; k < records.Len(); k++ {
				attrs := records.At(k).Attributes()
				for kName, v := range kvs {
					attrs.PutStr(kName, v)
				}
			}
		}
	}
	return ld, nil
}
