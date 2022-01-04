// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serialization // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/dynatraceexporter/serialization"

import (
	"fmt"

	"github.com/dynatrace-oss/dynatrace-metric-utils-go/metric/dimensions"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/common/ttlmap"
)

func SerializeMetric(logger *zap.Logger, prefix string, metric pdata.Metric, defaultDimensions, staticDimensions dimensions.NormalizedDimensionList, prev *ttlmap.TTLMap) ([]string, error) {
	var metricLines []string

	ce := logger.Check(zap.DebugLevel, "SerializeMetric")
	var points int

	switch metric.DataType() {
	case pdata.MetricDataTypeGauge:
		points = metric.Gauge().DataPoints().Len()
		for i := 0; i < metric.Gauge().DataPoints().Len(); i++ {
			dp := metric.Gauge().DataPoints().At(i)

			line, err := serializeGauge(
				metric.Name(),
				prefix,
				makeCombinedDimensions(dp.Attributes(), defaultDimensions, staticDimensions),
				dp,
			)

			if err != nil {
				return nil, err
			}

			if line != "" {
				metricLines = append(metricLines, line)
			}
		}
	case pdata.MetricDataTypeSum:
		points = metric.Sum().DataPoints().Len()
		for i := 0; i < metric.Sum().DataPoints().Len(); i++ {
			dp := metric.Sum().DataPoints().At(i)

			line, err := serializeSum(
				metric.Name(),
				prefix,
				makeCombinedDimensions(dp.Attributes(), defaultDimensions, staticDimensions),
				metric.Sum().AggregationTemporality(),
				dp,
				prev,
			)

			if err != nil {
				return nil, err
			}

			if line != "" {
				metricLines = append(metricLines, line)
			}
		}
	case pdata.MetricDataTypeHistogram:
		points = metric.Histogram().DataPoints().Len()
		for i := 0; i < metric.Histogram().DataPoints().Len(); i++ {
			dp := metric.Histogram().DataPoints().At(i)

			line, err := serializeHistogram(
				metric.Name(),
				prefix,
				makeCombinedDimensions(dp.Attributes(), defaultDimensions, staticDimensions),
				metric.Histogram().AggregationTemporality(),
				dp,
			)

			if err != nil {
				return nil, err
			}

			if line != "" {
				metricLines = append(metricLines, line)
			}
		}
	default:
		return nil, fmt.Errorf("metric type %s unsupported", metric.DataType().String())
	}

	if ce != nil {
		ce.Write(zap.String("DataType", metric.DataType().String()), zap.Int("points", points))
	}

	return metricLines, nil
}

func makeCombinedDimensions(labels pdata.AttributeMap, defaultDimensions, staticDimensions dimensions.NormalizedDimensionList) dimensions.NormalizedDimensionList {
	dimsFromLabels := []dimensions.Dimension{}

	labels.Range(func(k string, v pdata.AttributeValue) bool {
		dimsFromLabels = append(dimsFromLabels, dimensions.NewDimension(k, v.AsString()))
		return true
	})
	return dimensions.MergeLists(
		defaultDimensions,
		dimensions.NewNormalizedDimensionList(dimsFromLabels...),
		staticDimensions,
	)
}
