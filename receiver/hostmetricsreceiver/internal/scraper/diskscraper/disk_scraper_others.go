// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !windows
// +build !windows

package diskscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/diskscraper"

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/scrapererror"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filterset"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/diskscraper/internal/metadata"
)

const (
	standardMetricsLen = 5
	metricsLen         = standardMetricsLen + systemSpecificMetricsLen
)

// scraper for Disk Metrics
type scraper struct {
	config    *Config
	startTime pdata.Timestamp
	mb        *metadata.MetricsBuilder
	includeFS filterset.FilterSet
	excludeFS filterset.FilterSet

	// for mocking
	bootTime   func() (uint64, error)
	ioCounters func(names ...string) (map[string]disk.IOCountersStat, error)
}

// newDiskScraper creates a Disk Scraper
func newDiskScraper(_ context.Context, cfg *Config) (*scraper, error) {
	scraper := &scraper{config: cfg, bootTime: host.BootTime, ioCounters: disk.IOCounters}

	var err error

	if len(cfg.Include.Devices) > 0 {
		scraper.includeFS, err = filterset.CreateFilterSet(cfg.Include.Devices, &cfg.Include.Config)
		if err != nil {
			return nil, fmt.Errorf("error creating device include filters: %w", err)
		}
	}

	if len(cfg.Exclude.Devices) > 0 {
		scraper.excludeFS, err = filterset.CreateFilterSet(cfg.Exclude.Devices, &cfg.Exclude.Config)
		if err != nil {
			return nil, fmt.Errorf("error creating device exclude filters: %w", err)
		}
	}

	return scraper, nil
}

func (s *scraper) start(context.Context, component.Host) error {
	bootTime, err := s.bootTime()
	if err != nil {
		return err
	}

	s.startTime = pdata.Timestamp(bootTime * 1e9)
	s.mb = metadata.NewMetricsBuilder(s.config.Metrics, metadata.WithStartTime(s.startTime))
	return nil
}

func (s *scraper) scrape(_ context.Context) (pdata.Metrics, error) {
	md := pdata.NewMetrics()
	metrics := md.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty().Metrics()

	now := pdata.NewTimestampFromTime(time.Now())
	ioCounters, err := s.ioCounters()
	if err != nil {
		return md, scrapererror.NewPartialScrapeError(err, metricsLen)
	}

	// filter devices by name
	ioCounters = s.filterByDevice(ioCounters)

	if len(ioCounters) > 0 {
		metrics.EnsureCapacity(metricsLen)
		s.recordDiskIOMetric(now, ioCounters)
		s.recordDiskOperationsMetric(now, ioCounters)
		s.recordDiskIOTimeMetric(now, ioCounters)
		s.recordDiskOperationTimeMetric(now, ioCounters)
		s.recordDiskPendingOperationsMetric(now, ioCounters)
		s.recordSystemSpecificDataPoints(now, ioCounters)
		s.mb.Emit(metrics)
	}

	return md, nil
}

func (s *scraper) recordDiskIOMetric(now pdata.Timestamp, ioCounters map[string]disk.IOCountersStat) {
	for device, ioCounter := range ioCounters {
		s.mb.RecordSystemDiskIoDataPoint(now, int64(ioCounter.ReadBytes), device, metadata.AttributeDirection.Read)
		s.mb.RecordSystemDiskIoDataPoint(now, int64(ioCounter.WriteBytes), device, metadata.AttributeDirection.Write)
	}
}

func (s *scraper) recordDiskOperationsMetric(now pdata.Timestamp, ioCounters map[string]disk.IOCountersStat) {
	for device, ioCounter := range ioCounters {
		s.mb.RecordSystemDiskOperationsDataPoint(now, int64(ioCounter.ReadCount), device, metadata.AttributeDirection.Read)
		s.mb.RecordSystemDiskOperationsDataPoint(now, int64(ioCounter.WriteCount), device, metadata.AttributeDirection.Write)
	}
}

func (s *scraper) recordDiskIOTimeMetric(now pdata.Timestamp, ioCounters map[string]disk.IOCountersStat) {
	for device, ioCounter := range ioCounters {
		s.mb.RecordSystemDiskIoTimeDataPoint(now, float64(ioCounter.IoTime)/1e3, device)
	}
}

func (s *scraper) recordDiskOperationTimeMetric(now pdata.Timestamp, ioCounters map[string]disk.IOCountersStat) {
	for device, ioCounter := range ioCounters {
		s.mb.RecordSystemDiskOperationTimeDataPoint(now, float64(ioCounter.ReadTime)/1e3, device, metadata.AttributeDirection.Read)
		s.mb.RecordSystemDiskOperationTimeDataPoint(now, float64(ioCounter.WriteTime)/1e3, device, metadata.AttributeDirection.Write)
	}
}

func (s *scraper) recordDiskPendingOperationsMetric(now pdata.Timestamp, ioCounters map[string]disk.IOCountersStat) {
	for device, ioCounter := range ioCounters {
		s.mb.RecordSystemDiskPendingOperationsDataPoint(now, int64(ioCounter.IopsInProgress), device)
	}
}

func (s *scraper) filterByDevice(ioCounters map[string]disk.IOCountersStat) map[string]disk.IOCountersStat {
	if s.includeFS == nil && s.excludeFS == nil {
		return ioCounters
	}

	for device := range ioCounters {
		if !s.includeDevice(device) {
			delete(ioCounters, device)
		}
	}
	return ioCounters
}

func (s *scraper) includeDevice(deviceName string) bool {
	return (s.includeFS == nil || s.includeFS.Matches(deviceName)) &&
		(s.excludeFS == nil || !s.excludeFS.Matches(deviceName))
}
