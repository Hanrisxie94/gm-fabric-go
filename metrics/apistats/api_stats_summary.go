// Copyright 2017 Decipher Technology Studios LLC
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

package apistats

// APIEndpointStats represents stats for a single endpoint, or the total for all
type APIEndpointStats struct {
	Avg   float64 `json:"latency_ms.avg"`
	Count int64   `json:"latency_ms.count"`
	Max   int64   `json:"latency_ms.max"`
	Min   int64   `json:"latency_ms.min"`
	Sum   int64   `json:"latency_ms.sum"`
	P50   int64   `json:"latency_ms.p50"`
	P90   int64   `json:"latency_ms.p90"`
	P95   int64   `json:"latency_ms.p95"`
	P99   int64   `json:"latency_ms.p99"`
	P9990 int64   `json:"latency_ms.p9990"`
	P9999 int64   `json:"latency_ms.p9999"`

	Errors int32 `json:"errors.count"`

	InThroughput  int64 `json:"in_throughput"`
	OutThroughput int64 `json:"out_throughput"`

	Routes map[string]struct{}
}

// APIStatsSummary is stats on latency, etc in an API
// This is intended to be marshalled to JSON for external reporting
type APIStatsSummary struct {
	APIStats map[string]APIEndpointStats `json:"api"`
}

// EndpointStatsGetter provides access to latency statistics
type EndpointStatsGetter interface {

	// GetEndpointStats returns a snapshot of the current statistics
	GetEndpointStats() (map[string]APIEndpointStats, error)

	// GetCumulativeCounts returns cumulative counts of events
	GetCumulativeCounts() CumulativeCounts
}
