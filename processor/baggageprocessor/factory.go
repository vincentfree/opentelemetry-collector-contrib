// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package baggageprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/baggageprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/baggageprocessor/internal/metadata"
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}

// NewFactory returns a new factory for the Baggage processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, metadata.TracesStability),
		processor.WithLogs(createLogsProcessor, metadata.LogsStability),
		processor.WithMetrics(createMetricsProcessor, metadata.MetricsStability))
}

func createDefaultConfig() component.Config {
	return &Config{
		AttributePrefix:    "baggage.",
		MaxBaggageSize:     8192, // W3C specification default
		DropInvalidBaggage: false,
	}
}

func createTracesProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {
	oCfg := cfg.(*Config)
	bp := newBaggageProcessor(set.Logger, oCfg)

	return processorhelper.NewTraces(
		ctx,
		set,
		cfg,
		nextConsumer,
		bp.processTraces,
		processorhelper.WithCapabilities(processorCapabilities))
}

func createLogsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {
	oCfg := cfg.(*Config)
	bp := newBaggageProcessor(set.Logger, oCfg)

	return processorhelper.NewLogs(
		ctx,
		set,
		cfg,
		nextConsumer,
		bp.processLogs,
		processorhelper.WithCapabilities(processorCapabilities))
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	oCfg := cfg.(*Config)
	bp := newBaggageProcessor(set.Logger, oCfg)

	return processorhelper.NewMetrics(
		ctx,
		set,
		cfg,
		nextConsumer,
		bp.processMetrics,
		processorhelper.WithCapabilities(processorCapabilities))
}
