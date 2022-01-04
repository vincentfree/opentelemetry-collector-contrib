// Copyright 2020, OpenTelemetry Authors
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

package skywalkingexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/model/pdata"
	metricpb "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

func TestMetricDataToLogService(t *testing.T) {
	md := pdata.NewMetrics()
	md.ResourceMetrics().AppendEmpty() // Add an empty ResourceMetrics
	rm := md.ResourceMetrics().AppendEmpty()

	rm.Resource().Attributes().InsertString("labelB", "valueB")
	rm.Resource().Attributes().InsertString("labelA", "valueA")
	rm.Resource().Attributes().InsertString("a", "b")
	ilms := rm.InstrumentationLibraryMetrics()
	ilms.AppendEmpty() // Add an empty InstrumentationLibraryMetrics
	ilm := ilms.AppendEmpty()

	metrics := ilm.Metrics()

	badNameMetric := metrics.AppendEmpty()
	badNameMetric.SetName("")

	noneMetric := metrics.AppendEmpty()
	noneMetric.SetName("none")

	intGaugeMetric := metrics.AppendEmpty()
	intGaugeMetric.SetDataType(pdata.MetricDataTypeGauge)
	intGaugeMetric.SetName("int_gauge")
	intGauge := intGaugeMetric.Gauge()
	intGaugeDataPoints := intGauge.DataPoints()
	intGaugeDataPoint := intGaugeDataPoints.AppendEmpty()
	intGaugeDataPoint.Attributes().InsertString("innerLabel", "innerValue")
	intGaugeDataPoint.Attributes().InsertString("testa", "test")
	intGaugeDataPoint.SetIntVal(10)
	intGaugeDataPoint.SetTimestamp(pdata.Timestamp(100_000_000))

	doubleGaugeMetric := metrics.AppendEmpty()
	doubleGaugeMetric.SetDataType(pdata.MetricDataTypeGauge)
	doubleGaugeMetric.SetName("double_gauge")
	doubleGauge := doubleGaugeMetric.Gauge()
	doubleGaugeDataPoints := doubleGauge.DataPoints()
	doubleGaugeDataPoint := doubleGaugeDataPoints.AppendEmpty()
	doubleGaugeDataPoint.Attributes().InsertString("innerLabel", "innerValue")
	doubleGaugeDataPoint.SetDoubleVal(10.1)
	doubleGaugeDataPoint.SetTimestamp(pdata.Timestamp(100_000_000))

	intSumMetric := metrics.AppendEmpty()
	intSumMetric.SetDataType(pdata.MetricDataTypeSum)
	intSumMetric.SetName("int_sum")
	intSum := intSumMetric.Sum()
	intSumDataPoints := intSum.DataPoints()
	intSumDataPoint := intSumDataPoints.AppendEmpty()
	intSumDataPoint.Attributes().InsertString("innerLabel", "innerValue")
	intSumDataPoint.SetIntVal(11)
	intSumDataPoint.SetTimestamp(pdata.Timestamp(100_000_000))

	doubleSumMetric := metrics.AppendEmpty()
	doubleSumMetric.SetDataType(pdata.MetricDataTypeSum)
	doubleSumMetric.SetName("double_sum")
	doubleSum := doubleSumMetric.Sum()
	doubleSumDataPoints := doubleSum.DataPoints()
	doubleSumDataPoint := doubleSumDataPoints.AppendEmpty()
	doubleSumDataPoint.Attributes().InsertString("innerLabel", "innerValue")
	doubleSumDataPoint.SetDoubleVal(10.1)
	doubleSumDataPoint.SetTimestamp(pdata.Timestamp(100_000_000))

	doubleHistogramMetric := metrics.AppendEmpty()
	doubleHistogramMetric.SetDataType(pdata.MetricDataTypeHistogram)
	doubleHistogramMetric.SetName("double_$histogram")
	doubleHistogram := doubleHistogramMetric.Histogram()
	doubleHistogramDataPoints := doubleHistogram.DataPoints()
	doubleHistogramDataPoint := doubleHistogramDataPoints.AppendEmpty()
	doubleHistogramDataPoint.Attributes().InsertString("innerLabel", "innerValue")
	doubleHistogramDataPoint.Attributes().InsertString("innerLabelH", "innerValueH")
	doubleHistogramDataPoint.SetCount(5)
	doubleHistogramDataPoint.SetSum(10.1)
	doubleHistogramDataPoint.SetTimestamp(pdata.Timestamp(100_000_000))
	doubleHistogramDataPoint.SetBucketCounts([]uint64{1, 2, 2})
	doubleHistogramDataPoint.SetExplicitBounds([]float64{1, 2})

	doubleSummaryMetric := metrics.AppendEmpty()
	doubleSummaryMetric.SetDataType(pdata.MetricDataTypeSummary)
	doubleSummaryMetric.SetName("double-summary")
	doubleSummary := doubleSummaryMetric.Summary()
	doubleSummaryDataPoints := doubleSummary.DataPoints()
	doubleSummaryDataPoint := doubleSummaryDataPoints.AppendEmpty()
	doubleSummaryDataPoint.SetCount(2)
	doubleSummaryDataPoint.SetSum(10.1)
	doubleSummaryDataPoint.SetTimestamp(pdata.Timestamp(100_000_000))
	doubleSummaryDataPoint.Attributes().InsertString("innerLabel", "innerValue")
	doubleSummaryDataPoint.Attributes().InsertString("innerLabelS", "innerValueS")
	quantileVal := doubleSummaryDataPoint.QuantileValues().AppendEmpty()
	quantileVal.SetValue(10.2)
	quantileVal.SetQuantile(0.9)
	quantileVal2 := doubleSummaryDataPoint.QuantileValues().AppendEmpty()
	quantileVal2.SetValue(10.5)
	quantileVal2.SetQuantile(0.95)

	gotLogs := metricsRecordToMetricData(md)

	assert.Equal(t, 11, len(gotLogs.MeterData))

	for i, meterData := range gotLogs.MeterData {
		assert.Equal(t, "valueB", searchMetricTag("labelB", meterData))
		assert.Equal(t, "valueA", searchMetricTag("labelA", meterData))
		assert.Equal(t, "b", searchMetricTag("a", meterData))
		assert.Equal(t, "innerValue", searchMetricTag("innerLabel", meterData))
		assert.Equal(t, defaultServiceName, meterData.GetService())
		assert.Equal(t, defaultServiceInstance, meterData.GetServiceInstance())
		if i == 0 {
			assert.Equal(t, "int_gauge", meterData.GetSingleValue().GetName())
			assert.Equal(t, float64(10), meterData.GetSingleValue().GetValue())
			assert.Equal(t, int64(100), meterData.GetTimestamp())
			assert.Equal(t, "test", searchMetricTag("testa", meterData))
		} else if i == 1 {
			assert.Equal(t, "double_gauge", meterData.GetSingleValue().GetName())
			assert.Equal(t, 10.1, meterData.GetSingleValue().GetValue())
			assert.Equal(t, int64(100), meterData.GetTimestamp())
		} else if i == 2 {
			assert.Equal(t, "int_sum", meterData.GetSingleValue().GetName())
			assert.Equal(t, float64(11), meterData.GetSingleValue().GetValue())
			assert.Equal(t, int64(100), meterData.GetTimestamp())
		} else if i == 3 {
			assert.Equal(t, "double_sum", meterData.GetSingleValue().GetName())
			assert.Equal(t, 10.1, meterData.GetSingleValue().GetValue())
			assert.Equal(t, int64(100), meterData.GetTimestamp())
		} else if i == 4 {
			assert.Equal(t, "double_$histogram", meterData.GetHistogram().GetName())
			assert.Equal(t, 3, len(meterData.GetHistogram().GetValues()))
			assert.Equal(t, int64(1), meterData.GetHistogram().GetValues()[0].Count)
			assert.Equal(t, true, meterData.GetHistogram().GetValues()[0].IsNegativeInfinity)
			assert.Equal(t, int64(2), meterData.GetHistogram().GetValues()[1].Count)
			assert.Equal(t, false, meterData.GetHistogram().GetValues()[1].IsNegativeInfinity)
			assert.Equal(t, float64(1), meterData.GetHistogram().GetValues()[1].GetBucket())
			assert.Equal(t, int64(2), meterData.GetHistogram().GetValues()[2].Count)
			assert.Equal(t, false, meterData.GetHistogram().GetValues()[2].IsNegativeInfinity)
			assert.Equal(t, float64(2), meterData.GetHistogram().GetValues()[2].GetBucket())
			assert.Equal(t, int64(100), meterData.GetTimestamp())
			assert.Equal(t, "innerValueH", searchMetricTag("innerLabelH", meterData))

		} else if i == 5 {
			assert.Equal(t, "double_$histogram_sum", meterData.GetSingleValue().GetName())
			assert.Equal(t, 10.1, meterData.GetSingleValue().GetValue())
			assert.Equal(t, int64(100), meterData.GetTimestamp())
			assert.Equal(t, "innerValueH", searchMetricTag("innerLabelH", meterData))
		} else if i == 6 {
			assert.Equal(t, "double_$histogram_count", meterData.GetSingleValue().GetName())
			assert.Equal(t, float64(5), meterData.GetSingleValue().GetValue())
			assert.Equal(t, int64(100), meterData.GetTimestamp())
			assert.Equal(t, "innerValueH", searchMetricTag("innerLabelH", meterData))
		} else if i == 7 {
			assert.Equal(t, "double-summary", meterData.GetSingleValue().GetName())
			assert.Equal(t, 10.2, meterData.GetSingleValue().GetValue())
			assert.Equal(t, int64(100), meterData.GetTimestamp())
			assert.Equal(t, "innerValueS", searchMetricTag("innerLabelS", meterData))
			assert.Equal(t, "0.9", searchMetricTag("quantile", meterData))
		} else if i == 8 {
			assert.Equal(t, "double-summary", meterData.GetSingleValue().GetName())
			assert.Equal(t, 10.5, meterData.GetSingleValue().GetValue())
			assert.Equal(t, int64(100), meterData.GetTimestamp())
			assert.Equal(t, "innerValueS", searchMetricTag("innerLabelS", meterData))
			assert.Equal(t, "0.95", searchMetricTag("quantile", meterData))
		} else if i == 9 {
			assert.Equal(t, "double-summary_sum", meterData.GetSingleValue().GetName())
			assert.Equal(t, 10.1, meterData.GetSingleValue().GetValue())
			assert.Equal(t, int64(100), meterData.GetTimestamp())
			assert.Equal(t, "innerValueS", searchMetricTag("innerLabelS", meterData))
		} else if i == 10 {
			assert.Equal(t, "double-summary_count", meterData.GetSingleValue().GetName())
			assert.Equal(t, float64(2), meterData.GetSingleValue().GetValue())
			assert.Equal(t, int64(100), meterData.GetTimestamp())
			assert.Equal(t, "innerValueS", searchMetricTag("innerLabelS", meterData))
		}
	}
}

func searchMetricTag(name string, record *metricpb.MeterData) string {
	if _, ok := record.GetMetric().(*metricpb.MeterData_SingleValue); ok {
		for _, tag := range record.GetSingleValue().GetLabels() {
			if tag.Name == name {
				return tag.GetValue()
			}
		}
	}

	if _, ok := record.GetMetric().(*metricpb.MeterData_Histogram); ok {
		for _, tag := range record.GetHistogram().GetLabels() {
			if tag.Name == name {
				return tag.GetValue()
			}
		}
	}
	return ""
}
