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

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/deciphernow/gm-fabric-go/metrics/subject"
	"github.com/montanaflynn/stats"
)

// APIStatsEntry reports stats on an individual API call
type APIStatsEntry struct {
	RequestID     string
	Key           string
	Transport     subject.EventTransport
	HTTPStatus    int
	PrevRoute     string
	Err           error
	BeginTime     time.Time
	EndTime       time.Time
	InWireLength  int64
	OutWireLength int64
}

type APIStats struct {
	sync.Mutex
	Cache  *APIStatsCache
	Counts CumulativeCounts
}

// indices to percentiles
// not using iota here because thiese are fixed values
const (
	p50    = 0
	p90    = 1
	p95    = 2
	p99    = 3
	p9990  = 4
	p9999  = 5
	pCount = 6 // number of percentile values
)

func New(cacheSize int) *APIStats {
	return &APIStats{
		Cache:  NewAPIStatsCache(cacheSize),
		Counts: newCumulativeCounts(),
	}
}

func (st *APIStats) Store(entry APIStatsEntry) {
	st.Lock()
	defer st.Unlock()

	st.Cache.Store(entry)

	st.Counts.TotalEvents++
	st.Counts.TransportEvents[entry.Transport]++
	keyEvents, ok := st.Counts.KeyEvents[entry.Key]
	if !ok {
		keyEvents.StatusEvents = make(map[int]int64)
		keyEvents.StatusClassEvents = make(map[string]int64)
	}
	keyEvents.Events++
	if entry.Transport == subject.EventTransportHTTP ||
		entry.Transport == subject.EventTransportHTTPS {
		keyEvents.StatusEvents[entry.HTTPStatus]++
		keyEvents.StatusClassEvents[statusClass(entry.HTTPStatus)]++
	}
	st.Counts.KeyEvents[entry.Key] = keyEvents
}

// GetLatencyStats returns a summary of the stats currently available in the cache
func (st *APIStats) GetEndpointStats() (map[string]APIEndpointStats, error) {
	st.Lock()
	defer st.Unlock()

	apiStats, apiElapsedMS := st.accumulateEndpointStats()

	allEndpoints, allEndpointsElapsedMS, err := st.computeEndpointStats(apiStats, apiElapsedMS)
	if err != nil {
		return nil, errors.Wrapf(err, "computeEndpointStats")
	}

	if allEndpoints.Count > 0 {
		allEndpoints.Avg = float64(allEndpoints.Sum) / float64(allEndpoints.Count)
		percentileValues, err := computePercentiles(allEndpointsElapsedMS)
		if err != nil {
			return nil, errors.Wrapf(err, "computePercentiles")
		}
		allEndpoints.P50 = percentileValues[p50]
		allEndpoints.P90 = percentileValues[p90]
		allEndpoints.P95 = percentileValues[p95]
		allEndpoints.P99 = percentileValues[p99]
		allEndpoints.P9990 = percentileValues[p9990]
		allEndpoints.P9999 = percentileValues[p9999]
	}

	apiStats["all"] = allEndpoints

	return apiStats, nil
}

func (st *APIStats) accumulateEndpointStats() (
	map[string]APIEndpointStats,
	map[string][]float64,
) {
	apiStats := make(map[string]APIEndpointStats)
	apiElapsedMS := make(map[string][]float64)
	keyPrefix := ""

	// read through all cached transactions (trans) accumulating stats per
	// endpoint
	for trans := range st.Cache.Traverse() {

		// we expect the key to be of the form 'route/...' or 'function/...'
		if keyPrefix == "" {
			s := strings.Split(trans.Key, "/")
			if len(s) > 0 {
				keyPrefix = s[0]
			}
		}

		endpoint := apiStats[trans.Key]
		if endpoint.Routes == nil {
			endpoint.Routes = make(map[string]struct{})
		}
		if trans.PrevRoute != "" {
			endpoint.Routes[trans.PrevRoute] = struct{}{}
		}
		endpointElapsedMS := apiElapsedMS[trans.Key]

		elapsed := trans.EndTime.Sub(trans.BeginTime)
		elapsedMS := duration2ms(elapsed)

		endpointElapsedMS = append(endpointElapsedMS, float64(elapsedMS))

		endpoint.Count++
		endpoint.Sum += elapsedMS

		if endpoint.Min == 0 || elapsedMS < endpoint.Min {
			endpoint.Min = elapsedMS
		}
		if elapsedMS > endpoint.Max {
			endpoint.Max = elapsedMS
		}
		if trans.Err != nil {
			endpoint.Errors++
		}
		endpoint.InThroughput += trans.InWireLength
		endpoint.OutThroughput += trans.OutWireLength

		apiStats[trans.Key] = endpoint
		apiElapsedMS[trans.Key] = endpointElapsedMS
	}

	return apiStats, apiElapsedMS
}

func (st *APIStats) computeEndpointStats(
	apiStats map[string]APIEndpointStats,
	apiElapsedMS map[string][]float64,
) (APIEndpointStats, []float64, error) {
	var allEndpoints APIEndpointStats
	var allEndpointsElapsedMS []float64

	// read through accumulated endpoint stats computing statistical values
	// (AVG, etc.), and accumulating total for all endpoints

	for name := range apiStats {
		endpoint := apiStats[name]
		endpointElapsedMS := apiElapsedMS[name]

		allEndpoints.Count += endpoint.Count
		if endpoint.Count > 0 {
			endpoint.Avg = float64(endpoint.Sum) / float64(endpoint.Count)
		}

		if endpoint.Max > allEndpoints.Max {
			allEndpoints.Max = endpoint.Max
		}
		if allEndpoints.Min == 0 || endpoint.Min < allEndpoints.Min {
			allEndpoints.Min = endpoint.Min
		}
		allEndpoints.Sum += endpoint.Sum

		if endpoint.Count > 0 {
			percentileValues, err := computePercentiles(endpointElapsedMS)
			if err != nil {
				return APIEndpointStats{}, nil, errors.Wrapf(err, "computePercentiles")
			}
			endpoint.P50 = percentileValues[p50]
			endpoint.P90 = percentileValues[p90]
			endpoint.P95 = percentileValues[p95]
			endpoint.P99 = percentileValues[p99]
			endpoint.P9990 = percentileValues[p9990]
			endpoint.P9999 = percentileValues[p9999]
		}

		allEndpointsElapsedMS = append(allEndpointsElapsedMS, endpointElapsedMS...)

		allEndpoints.InThroughput += endpoint.InThroughput
		allEndpoints.OutThroughput += endpoint.OutThroughput

		allEndpoints.Errors += endpoint.Errors

		apiStats[name] = endpoint
	}

	return allEndpoints, allEndpointsElapsedMS, nil
}

// GetCumulativeCounts returns cumulative counts of events
func (st *APIStats) GetCumulativeCounts() CumulativeCounts {
	st.Lock()
	defer st.Unlock()

	return copyCumulativeCounts(st.Counts)
}

func duration2ms(d time.Duration) int64 {
	const nsPerMs = 1000000

	return d.Nanoseconds() / nsPerMs
}

func computePercentiles(elapsedMS []float64) ([]int64, error) {
	var err error

	targets := []float64{50, 90, 95, 99, 99.9, 99.99}
	values := make([]int64, pCount)

	for i, target := range targets {
		var p float64
		p, err = stats.PercentileNearestRank(elapsedMS, target)
		if err != nil {
			return nil, errors.Wrapf(err, "PercentileNearestRank")
		}
		values[i] = int64(p)
	}

	return values, nil
}

// statusClass returns a string of the form NXX for HTTP Status, e.g. 2XX
func statusClass(status int) string {
	return fmt.Sprintf("%dXX", status/100)
}
