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
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/deciphernow/gm-fabric-go/metrics/subject"
	"github.com/montanaflynn/stats"
)

// APIStatsEntry reports stats on an individual API call
type APIStatsEntry struct {
	RequestID  string
	Key        string
	Transport  subject.EventTransport
	HTTPStatus int
	PrevRoute  string
	Err        error

	// BeginTime is the earliest point that we can store a timestamp
	BeginTime time.Time

	// RequestTime is the time when an HTTP request is available, with all headers
	// received
	// Latency = RequestTime - BeginTime
	RequestTime time.Time

	// ResponseTime is the time when an HTTP response is ready to send
	ResponseTime time.Time

	// EndTime is the time when the transaction is completely ended
	// We may not capture this time for long running transactions
	EndTime time.Time

	// InWireLength is the amount of data received from the request body
	// at InCaptureTime
	// WireLength is the length of data on wire (compressed, signed, encrypted).
	// InThroughput = InWireLength / (InCaptureTime - RequestTime)
	InWireLength int64

	// InCaptureTime is the time when InWireLength is captured
	// This may be the same as EndTime for short transactions
	// but may be earlier for long running transactions.
	InCaptureTime time.Time

	// OutWireLength is the amount of data sent from the response body
	// at OutCaptureTime
	// WireLength is the length of data on wire (compressed, signed, encrypted).
	// OutThroughput = OutWireLength / (OutCaptureTime - ResponseTime)
	OutWireLength int64

	// OutCaptureTime is the time when OutWireLength is captured
	// This may be the same as EndTime for short transactions
	// but may be earlier for long running transactions.
	OutCaptureTime time.Time
}

type APIStats struct {
	sync.Mutex
	Cache  *APIStatsCache
	Counts CumulativeCounts
}

type throughputAccum struct {
	receivedSecs  int64
	bytesReceived int64
	sentSecs      int64
	bytesSent     int64
}

type endpointResultType struct {
	throughputAccum
	apiStats       map[string]APIEndpointStats
	latencySamples map[string][]float64
}

type allEndpointsResultType struct {
	allEndpoints               APIEndpointStats
	allEndpointsLatencySamples []float64
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

	endpointResult := st.accumulateEndpointStats()

	// computeEndpointStats modifies endpointResult
	allEndpointsResult, err := st.computeEndpointStats(&endpointResult)
	if err != nil {
		return nil, errors.Wrapf(err, "computeEndpointStats")
	}

	if allEndpointsResult.allEndpoints.Count > 0 {
		allEndpointsResult.allEndpoints.Avg =
			float64(allEndpointsResult.allEndpoints.Sum) / float64(allEndpointsResult.allEndpoints.Count)
		percentileValues, err :=
			computePercentiles(allEndpointsResult.allEndpointsLatencySamples)
		if err != nil {
			return nil, errors.Wrapf(err, "computePercentiles")
		}
		allEndpointsResult.allEndpoints.P50 = percentileValues[p50]
		allEndpointsResult.allEndpoints.P90 = percentileValues[p90]
		allEndpointsResult.allEndpoints.P95 = percentileValues[p95]
		allEndpointsResult.allEndpoints.P99 = percentileValues[p99]
		allEndpointsResult.allEndpoints.P9990 = percentileValues[p9990]
		allEndpointsResult.allEndpoints.P9999 = percentileValues[p9999]

		if endpointResult.receivedSecs > 0 {
			allEndpointsResult.allEndpoints.InThroughput =
				endpointResult.bytesReceived / endpointResult.receivedSecs
		}

		if endpointResult.sentSecs > 0 {
			allEndpointsResult.allEndpoints.OutThroughput =
				endpointResult.bytesSent / endpointResult.sentSecs
		}
	}

	endpointResult.apiStats["all"] = allEndpointsResult.allEndpoints

	return endpointResult.apiStats, nil
}

func (st *APIStats) accumulateEndpointStats() endpointResultType {
	endpointResult := endpointResultType{
		apiStats:       make(map[string]APIEndpointStats),
		latencySamples: make(map[string][]float64),
	}
	endpointAccumMap := make(map[string]throughputAccum)

	// read through all cached transactions (trans) accumulating stats per
	// endpoint
	for trans := range st.Cache.Traverse() {

		endpoint := endpointResult.apiStats[trans.Key]
		if endpoint.Routes == nil {
			endpoint.Routes = make(map[string]struct{})
		}
		if trans.PrevRoute != "" {
			endpoint.Routes[trans.PrevRoute] = struct{}{}
		}
		endpointLatencySamples := endpointResult.latencySamples[trans.Key]
		endpointAccum := endpointAccumMap[trans.Key]

		latencyDuration := trans.RequestTime.Sub(trans.BeginTime)
		latency := duration2ms(latencyDuration)

		endpointLatencySamples = append(endpointLatencySamples, float64(latency))

		endpoint.Count++
		endpoint.Sum += latency

		if endpoint.Min == 0 || latency < endpoint.Min {
			endpoint.Min = latency
		}
		if latency > endpoint.Max {
			endpoint.Max = latency
		}
		if trans.Err != nil {
			endpoint.Errors++
		}

		if (!trans.InCaptureTime.IsZero()) && (!trans.RequestTime.IsZero()) {
			requestSecs := int64(trans.InCaptureTime.Sub(trans.RequestTime).Seconds())
			if requestSecs > 0 {
				endpointAccum.receivedSecs += requestSecs
				endpointAccum.bytesReceived += trans.InWireLength
			}
		}

		if (!trans.OutCaptureTime.IsZero()) && (!trans.ResponseTime.IsZero()) {
			responseSecs := int64(trans.OutCaptureTime.Sub(trans.ResponseTime).Seconds())
			if responseSecs > 0 {
				endpointAccum.sentSecs += responseSecs
				endpointAccum.bytesSent += trans.OutWireLength
			}
		}

		endpointAccumMap[trans.Key] = endpointAccum
		endpointResult.apiStats[trans.Key] = endpoint
		endpointResult.latencySamples[trans.Key] = endpointLatencySamples
	}

	for key := range endpointAccumMap {
		endpoint := endpointResult.apiStats[key]
		endpointAccum := endpointAccumMap[key]

		if endpointAccum.receivedSecs > 0 {
			endpoint.InThroughput = endpointAccum.bytesReceived / endpointAccum.receivedSecs
			endpointResult.receivedSecs += endpointAccum.receivedSecs
			endpointResult.bytesReceived += endpointAccum.bytesReceived
		}

		if endpointAccum.sentSecs > 0 {
			endpoint.OutThroughput = endpointAccum.bytesSent / endpointAccum.sentSecs
			endpointResult.sentSecs += endpointAccum.sentSecs
			endpointResult.bytesSent += endpointAccum.bytesSent
		}

		endpointResult.apiStats[key] = endpoint
	}

	return endpointResult
}

func (st *APIStats) computeEndpointStats(
	endpointResult *endpointResultType,
) (allEndpointsResultType, error) {
	var allEndpointsResult allEndpointsResultType

	// read through accumulated endpoint stats computing statistical values
	// (AVG, etc.), and accumulating total for all endpoints

	for name := range endpointResult.apiStats {
		endpoint := endpointResult.apiStats[name]
		endpointLatencySamples := endpointResult.latencySamples[name]

		allEndpointsResult.allEndpoints.Count += endpoint.Count
		if endpoint.Count > 0 {
			endpoint.Avg = float64(endpoint.Sum) / float64(endpoint.Count)
		}

		if endpoint.Max > allEndpointsResult.allEndpoints.Max {
			allEndpointsResult.allEndpoints.Max = endpoint.Max
		}
		if allEndpointsResult.allEndpoints.Min == 0 || endpoint.Min < allEndpointsResult.allEndpoints.Min {
			allEndpointsResult.allEndpoints.Min = endpoint.Min
		}
		allEndpointsResult.allEndpoints.Sum += endpoint.Sum

		if endpoint.Count > 0 {
			percentileValues, err := computePercentiles(endpointLatencySamples)
			if err != nil {
				return allEndpointsResultType{}, errors.Wrapf(err, "computePercentiles")
			}
			endpoint.P50 = percentileValues[p50]
			endpoint.P90 = percentileValues[p90]
			endpoint.P95 = percentileValues[p95]
			endpoint.P99 = percentileValues[p99]
			endpoint.P9990 = percentileValues[p9990]
			endpoint.P9999 = percentileValues[p9999]
		}

		allEndpointsResult.allEndpointsLatencySamples =
			append(allEndpointsResult.allEndpointsLatencySamples, endpointLatencySamples...)

		allEndpointsResult.allEndpoints.Errors += endpoint.Errors

		endpointResult.apiStats[name] = endpoint
	}

	return allEndpointsResult, nil
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
