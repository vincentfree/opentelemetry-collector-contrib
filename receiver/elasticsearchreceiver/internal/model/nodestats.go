// Copyright  The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver/internal/model"

// NodeStats represents a response from elasticsearch's /_nodes/stats endpoint.
// The struct is not exhaustive; It does not provide all values returned by elasticsearch,
// only the ones relevant to the metrics retrieved by the scraper.
type NodeStats struct {
	ClusterName string                        `json:"cluster_name"`
	Nodes       map[string]NodeStatsNodesInfo `json:"nodes"`
}

type NodeStatsNodesInfo struct {
	TimestampMsSinceEpoch int64                      `json:"timestamp"`
	Name                  string                     `json:"name"`
	Indices               NodeStatsNodesInfoIndices  `json:"indices"`
	ProcessStats          ProcessStats               `json:"process"`
	JVMInfo               JVMInfo                    `json:"jvm"`
	ThreadPoolInfo        map[string]ThreadPoolStats `json:"thread_pool"`
	TransportStats        TransportStats             `json:"transport"`
	HTTPStats             HTTPStats                  `json:"http"`
	FS                    FSStats                    `json:"fs"`
}

type NodeStatsNodesInfoIndices struct {
	StoreInfo          StoreInfo           `json:"store"`
	DocumentStats      DocumentStats       `json:"docs"`
	IndexingOperations IndexingOperations  `json:"indexing"`
	GetOperation       GetOperation        `json:"get"`
	SearchOperations   SearchOperations    `json:"search"`
	MergeOperations    BasicIndexOperation `json:"merges"`
	RefreshOperations  BasicIndexOperation `json:"refresh"`
	FlushOperations    BasicIndexOperation `json:"flush"`
	WarmerOperations   BasicIndexOperation `json:"warmer"`
	QueryCache         BasicCacheInfo      `json:"query_cache"`
	FieldDataCache     BasicCacheInfo      `json:"fielddata"`
}

type StoreInfo struct {
	SizeInBy int64 `json:"size_in_bytes"`
}

type BasicIndexOperation struct {
	Total         int64 `json:"total"`
	TotalTimeInMs int64 `json:"total_time_in_millis"`
}

type IndexingOperations struct {
	IndexTotal     int64 `json:"index_total"`
	IndexTimeInMs  int64 `json:"index_time_in_millis"`
	DeleteTotal    int64 `json:"delete_total"`
	DeleteTimeInMs int64 `json:"delete_time_in_millis"`
}

type GetOperation struct {
	Total         int64 `json:"total"`
	TotalTimeInMs int64 `json:"time_in_millis"`
}

type SearchOperations struct {
	QueryTotal      int64 `json:"query_total"`
	QueryTimeInMs   int64 `json:"query_time_in_millis"`
	FetchTotal      int64 `json:"fetch_total"`
	FetchTimeInMs   int64 `json:"fetch_time_in_millis"`
	ScrollTotal     int64 `json:"scroll_total"`
	ScrollTimeInMs  int64 `json:"scroll_time_in_millis"`
	SuggestTotal    int64 `json:"suggest_total"`
	SuggestTimeInMs int64 `json:"suggest_time_in_millis"`
}

type DocumentStats struct {
	ActiveCount  int64 `json:"count"`
	DeletedCount int64 `json:"deleted"`
}

type BasicCacheInfo struct {
	Evictions      int64 `json:"evictions"`
	MemorySizeInBy int64 `json:"memory_size_in_bytes"`
}

type JVMInfo struct {
	UptimeInMs    int64         `json:"uptime_in_millis"`
	JVMMemoryInfo JVMMemoryInfo `json:"mem"`
	JVMThreadInfo JVMThreadInfo `json:"threads"`
	JVMGCInfo     JVMGCInfo     `json:"gc"`
}

type JVMMemoryInfo struct {
	HeapUsedInBy    int64 `json:"heap_used_in_bytes"`
	NonHeapUsedInBy int64 `json:"non_heap_used_in_bytes"`
	MaxHeapInBy     int64 `json:"heap_max_in_bytes"`
}

type JVMThreadInfo struct {
	PeakCount int64 `json:"peak_count"`
	Count     int64 `json:"count"`
}

type JVMGCInfo struct {
	Collectors JVMCollectors `json:"collectors"`
}

type JVMCollectors struct {
	Young BasicJVMCollectorInfo `json:"young"`
	Old   BasicJVMCollectorInfo `json:"old"`
}

type BasicJVMCollectorInfo struct {
	CollectionCount        int64 `json:"collection_count"`
	CollectionTimeInMillis int64 `json:"collection_time_in_millis"`
}

type ThreadPoolStats struct {
	TotalThreads   int64 `json:"threads"`
	ActiveThreads  int64 `json:"active"`
	QueuedTasks    int64 `json:"queue"`
	CompletedTasks int64 `json:"completed"`
	RejectedTasks  int64 `json:"rejected"`
}

type ProcessStats struct {
	OpenFileDescriptorsCount int64 `json:"open_file_descriptors"`
}

type TransportStats struct {
	OpenConnections int64 `json:"server_open"`
	ReceivedBytes   int64 `json:"rx_size_in_bytes"`
	SentBytes       int64 `json:"tx_size_in_bytes"`
}

type HTTPStats struct {
	OpenConnections int64 `json:"current_open"`
}

type FSStats struct {
	Total   FSTotalStats `json:"total"`
	IOStats *IOStats     `json:"io_stats,omitempty"`
}

type FSTotalStats struct {
	AvailableBytes int64 `json:"available_in_bytes"`
	TotalBytes     int64 `json:"total_in_bytes"`
}

type IOStats struct {
	Total IOStatsTotal `json:"total"`
}

type IOStatsTotal struct {
	ReadOperations  int64 `json:"read_operations"`
	WriteOperations int64 `json:"write_operations"`
}
