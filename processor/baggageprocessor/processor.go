// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package baggageprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/baggageprocessor"

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/baggage"
	"go.uber.org/zap"
)

// baggageProcessor implements the baggage processor
type baggageProcessor struct {
	logger *zap.Logger
	config *Config
}

// newBaggageProcessor creates a new baggage processor
func newBaggageProcessor(logger *zap.Logger, config *Config) *baggageProcessor {
	return &baggageProcessor{
		logger: logger,
		config: config,
	}
}

// processTraces processes trace data
func (bp *baggageProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	resourceSpans := td.ResourceSpans()
	for i := 0; i < resourceSpans.Len(); i++ {
		rs := resourceSpans.At(i)
		scopeSpans := rs.ScopeSpans()
		for j := 0; j < scopeSpans.Len(); j++ {
			ss := scopeSpans.At(j)
			spans := ss.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				if err := bp.processAttributes(ctx, span.Attributes()); err != nil {
					bp.logger.Error("Failed to process baggage for span", zap.Error(err))
					if !bp.config.DropInvalidBaggage {
						return td, err
					}
				}
			}
		}
	}
	return td, nil
}

// processLogs processes log data
func (bp *baggageProcessor) processLogs(ctx context.Context, ld plog.Logs) (plog.Logs, error) {
	resourceLogs := ld.ResourceLogs()
	for i := 0; i < resourceLogs.Len(); i++ {
		rl := resourceLogs.At(i)
		scopeLogs := rl.ScopeLogs()
		for j := 0; j < scopeLogs.Len(); j++ {
			sl := scopeLogs.At(j)
			logRecords := sl.LogRecords()
			for k := 0; k < logRecords.Len(); k++ {
				logRecord := logRecords.At(k)
				if err := bp.processAttributes(ctx, logRecord.Attributes()); err != nil {
					bp.logger.Error("Failed to process baggage for log record", zap.Error(err))
					if !bp.config.DropInvalidBaggage {
						return ld, err
					}
				}
			}
		}
	}
	return ld, nil
}

// processMetrics processes metric data
func (bp *baggageProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	resourceMetrics := md.ResourceMetrics()
	for i := 0; i < resourceMetrics.Len(); i++ {
		rm := resourceMetrics.At(i)
		scopeMetrics := rm.ScopeMetrics()
		for j := 0; j < scopeMetrics.Len(); j++ {
			sm := scopeMetrics.At(j)
			metrics := sm.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)
				if err := bp.processMetricDataPoints(ctx, metric); err != nil {
					bp.logger.Error("Failed to process baggage for metric", zap.Error(err))
					if !bp.config.DropInvalidBaggage {
						return md, err
					}
				}
			}
		}
	}
	return md, nil
}

// processMetricDataPoints processes data points in a metric
func (bp *baggageProcessor) processMetricDataPoints(ctx context.Context, metric pmetric.Metric) error {
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		dps := metric.Gauge().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			if err := bp.processAttributes(ctx, dps.At(i).Attributes()); err != nil {
				return err
			}
		}
	case pmetric.MetricTypeSum:
		dps := metric.Sum().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			if err := bp.processAttributes(ctx, dps.At(i).Attributes()); err != nil {
				return err
			}
		}
	case pmetric.MetricTypeHistogram:
		dps := metric.Histogram().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			if err := bp.processAttributes(ctx, dps.At(i).Attributes()); err != nil {
				return err
			}
		}
	case pmetric.MetricTypeExponentialHistogram:
		dps := metric.ExponentialHistogram().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			if err := bp.processAttributes(ctx, dps.At(i).Attributes()); err != nil {
				return err
			}
		}
	case pmetric.MetricTypeSummary:
		dps := metric.Summary().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			if err := bp.processAttributes(ctx, dps.At(i).Attributes()); err != nil {
				return err
			}
		}
	}
	return nil
}

// processAttributes processes attributes according to the configured actions
func (bp *baggageProcessor) processAttributes(ctx context.Context, attrs pcommon.Map) error {
	// Get current baggage from context
	currentBaggage := baggage.FromContext(ctx)

	// Process each action
	for _, action := range bp.config.Actions {
		if err := bp.processAction(ctx, action, attrs, &currentBaggage); err != nil {
			return fmt.Errorf("failed to process action for key %s: %w", action.Key, err)
		}
	}

	return nil
}

// processAction processes a single baggage action
func (bp *baggageProcessor) processAction(_ context.Context, action Action, attrs pcommon.Map, currentBaggage *baggage.Baggage) error {
	switch action.Action {
	case Extract:
		return bp.extractBaggage(action, attrs, *currentBaggage)
	case Inject:
		return bp.injectBaggage(action, attrs, currentBaggage)
	case Update:
		return bp.updateBaggage(action, attrs, currentBaggage, false)
	case Upsert:
		return bp.updateBaggage(action, attrs, currentBaggage, true)
	case Delete:
		return bp.deleteBaggage(action, currentBaggage)
	default:
		return fmt.Errorf("unknown action type: %s", action.Action)
	}
}

// extractBaggage extracts baggage entries and adds them as attributes
func (bp *baggageProcessor) extractBaggage(action Action, attrs pcommon.Map, currentBaggage baggage.Baggage) error {
	member := currentBaggage.Member(action.Key)
	if member.Key() == "" {
		// Baggage key not found, skip
		return nil
	}

	// Determine the attribute name
	attrName := action.ToAttribute
	if attrName == "" {
		attrName = bp.config.AttributePrefix + action.Key
	}

	// Set the attribute value
	attrs.PutStr(attrName, member.Value())

	// If there are properties, add them as separate attributes
	properties := member.Properties()
	if len(properties) > 0 {
		var propStrings []string
		for _, prop := range properties {
			propStrings = append(propStrings, prop.String())
		}
		attrs.PutStr(attrName+"_properties", strings.Join(propStrings, ";"))
	}

	return nil
}

// injectBaggage injects attributes as baggage entries
func (bp *baggageProcessor) injectBaggage(action Action, attrs pcommon.Map, currentBaggage *baggage.Baggage) error {
	var value string

	// Get the value to inject
	if action.Value != "" {
		value = action.Value
	} else if action.FromAttribute != "" {
		attrVal, exists := attrs.Get(action.FromAttribute)
		if !exists {
			// Attribute not found, skip
			return nil
		}
		value = attrVal.AsString()
	}

	// Create baggage member with properties
	var member baggage.Member
	var err error

	if len(action.Properties) > 0 {
		// Build properties
		var props []baggage.Property
		for key, val := range action.Properties {
			prop, propErr := baggage.NewKeyValueProperty(key, val)
			if propErr != nil {
				if bp.config.DropInvalidBaggage {
					bp.logger.Warn("Invalid baggage property, skipping",
						zap.String("key", key), zap.String("value", val), zap.Error(propErr))
					continue
				}
				return fmt.Errorf("invalid baggage property %s=%s: %w", key, val, propErr)
			}
			props = append(props, prop)
		}
		member, err = baggage.NewMember(action.Key, value, props...)
	} else {
		member, err = baggage.NewMember(action.Key, value)
	}

	if err != nil {
		if bp.config.DropInvalidBaggage {
			bp.logger.Warn("Invalid baggage member, skipping",
				zap.String("key", action.Key), zap.String("value", value), zap.Error(err))
			return nil
		}
		return fmt.Errorf("failed to create baggage member: %w", err)
	}

	// Add to baggage
	newBaggage, err := currentBaggage.SetMember(member)
	if err != nil {
		if bp.config.DropInvalidBaggage {
			bp.logger.Warn("Failed to set baggage member, skipping",
				zap.String("key", action.Key), zap.Error(err))
			return nil
		}
		return fmt.Errorf("failed to set baggage member: %w", err)
	}

	// Check baggage size if configured
	if bp.config.MaxBaggageSize > 0 {
		baggageStr := newBaggage.String()
		if len(baggageStr) > bp.config.MaxBaggageSize {
			if bp.config.DropInvalidBaggage {
				bp.logger.Warn("Baggage size exceeds limit, skipping",
					zap.Int("size", len(baggageStr)), zap.Int("limit", bp.config.MaxBaggageSize))
				return nil
			}
			return fmt.Errorf("baggage size %d exceeds limit %d", len(baggageStr), bp.config.MaxBaggageSize)
		}
	}

	*currentBaggage = newBaggage
	return nil
}

// updateBaggage updates existing baggage entries
func (bp *baggageProcessor) updateBaggage(action Action, attrs pcommon.Map, currentBaggage *baggage.Baggage, upsert bool) error {
	// Check if the key exists
	member := currentBaggage.Member(action.Key)
	if member.Key() == "" && !upsert {
		// Key doesn't exist and we're not upserting, skip
		return nil
	}

	// Use inject logic for update/upsert
	return bp.injectBaggage(action, attrs, currentBaggage)
}

// deleteBaggage removes baggage entries
func (bp *baggageProcessor) deleteBaggage(action Action, currentBaggage *baggage.Baggage) error {
	newBaggage := currentBaggage.DeleteMember(action.Key)
	*currentBaggage = newBaggage
	return nil
}
