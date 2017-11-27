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

package grpcobserver

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/shirou/gopsutil/cpu"

	"github.com/deciphernow/gm-fabric-go/metrics/flatjson"
	"github.com/deciphernow/gm-fabric-go/metrics/memvalues"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
)

// Report implements the Reporter interface it is called by the metrics server
func (obs *GRPCObserver) Report(jWriter *flatjson.Writer) error {
	var err error
	var summary APIStatsSummary
	var counts cumulativeCounts
	transportLabels := map[subject.EventTransport]string{
		subject.EventTransportHTTP:       "HTTP",
		subject.EventTransportHTTPS:      "HTTPS",
		subject.EventTransportRPC:        "RPC",
		subject.EventTransportRPCWithTLS: "RPC_TLS",
	}

	if summary.APIStats, counts, err = obs.getAPIStats(); err != nil {
		return err
	}

	err = jWriter.Write(fmt.Sprintf("%s/%s", "Total", "requests"), counts.totalEvents)
	if err != nil {
		return err
	}

	for _, transport := range []subject.EventTransport{
		subject.EventTransportHTTP,
		subject.EventTransportHTTPS,
		subject.EventTransportRPC,
		subject.EventTransportRPCWithTLS,
	} {
		err = jWriter.Write(
			fmt.Sprintf("%s/%s", transportLabels[transport], "requests"),
			counts.transportEvents[transport],
		)
		if err != nil {
			return err
		}
	}

	var allEvents int64
	for path := range summary.APIStats {
		if path != "all" {
			allEvents += counts.keyEvents[path]
		}
	}

	for path, value := range summary.APIStats {
		var count int64
		if path == "all" {
			count = allEvents
		} else {
			count = counts.keyEvents[path]
		}
		err = jWriter.Write(fmt.Sprintf("%s/%s", path, "requests"), count)
		if err != nil {
			return err
		}

		var routes string
		for route := range value.Routes {
			if len(routes) == 0 {
				routes = route
			} else {
				routes = fmt.Sprintf("%s:%s", routes, route)
			}
		}
		err = jWriter.Write(fmt.Sprintf("%s/%s", path, "routes"), routes)
		if err != nil {
			return err
		}

		for _, x := range []struct {
			label string
			val   interface{}
		}{
			{"latency_ms.avg", value.Avg},
			{"latency_ms.count", value.Count},
			{"latency_ms.max", value.Max},
			{"latency_ms.min", value.Min},
			{"latency_ms.sum", value.Sum},
			{"latency_ms.p50", value.P50},
			{"latency_ms.p90", value.P90},
			{"latency_ms.p95", value.P95},
			{"latency_ms.p99", value.P99},
			{"latency_ms.p9990", value.P9990},
			{"latency_ms.p9999", value.P9999},
			{"errors.count", value.Errors},
			{"in_throughput", value.InThroughput},
			{"out_throughput", value.OutThroughput},
		} {
			err = jWriter.Write(fmt.Sprintf("%s/%s", path, x.label), x.val)
			if err != nil {
				return err
			}
		}
	}

	memValues, err := memvalues.GetMemValues()
	if err != nil {
		return err
	}

	var cpuPercent []float64
	cpuPercent, err = cpu.Percent(time.Second, false)
	if err != nil {
		return fmt.Errorf("cpu.Percent failed: %v", err)
	}

	for _, x := range []struct {
		key   string
		value interface{}
	}{
		{"system/start_time", duration2ms(time.Duration(obs.startTime.UnixNano()))},
		{"system/cpu.pct", cpuPercent[0]},
		{"system/cpu_cores", runtime.NumCPU()},
		{"os", runtime.GOOS},
		{"os_arch", runtime.GOARCH},
		{"system/memory/available", memValues.SystemMemoryAvailable},
		{"system/memory/used", memValues.SystemMemoryUsed},
		{"system/memory/used_percent", memValues.SystemMemoryUsedPercent},
		{"process/memory/used", memValues.ProcessMemoryUsed},
	} {
		if err = jWriter.Write(x.key, x.value); err != nil {
			return err
		}
	}

	return nil
}

// GetLatencyStats returns a snapshot of the current latency statistics
func (obs *GRPCObserver) GetLatencyStats() (map[string]APIEndpointStats, error) {
	stats, _, err := obs.getAPIStats()
	return stats, err
}

func (obs *GRPCObserver) getAPIStats() (map[string]APIEndpointStats, cumulativeCounts, error) {
	obs.lock.Lock()
	defer obs.lock.Unlock()

	apiStats := make(map[string]APIEndpointStats)
	apiElapsedMS := make(map[string][]float64)

	keyPrefix := ""

	// read through all cached transactions (trans) accumulating stats per
	// endpoint
	for trans := range obs.cache.traverse() {

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
				return nil, cumulativeCounts{}, fmt.Errorf("computePercentiles failed: %v", err)
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

	if allEndpoints.Count > 0 {
		allEndpoints.Avg = float64(allEndpoints.Sum) / float64(allEndpoints.Count)
		percentileValues, err := computePercentiles(allEndpointsElapsedMS)
		if err != nil {
			return nil, cumulativeCounts{}, fmt.Errorf("computePercentiles failed: %v", err)
		}
		allEndpoints.P50 = percentileValues[p50]
		allEndpoints.P90 = percentileValues[p90]
		allEndpoints.P95 = percentileValues[p95]
		allEndpoints.P99 = percentileValues[p99]
		allEndpoints.P9990 = percentileValues[p9990]
		allEndpoints.P9999 = percentileValues[p9999]
	}

	apiStats["all"] = allEndpoints

	return apiStats, copyCumulativeCounts(obs.cumulativeCounts), nil
}

func duration2ms(d time.Duration) int64 {
	const nsPerMs = 1000000

	return int64(d.Nanoseconds() / nsPerMs)
}

func computePercentiles(elapsedMS []float64) ([]int64, error) {
	var err error

	targets := []float64{50, 90, 95, 99, 99.9, 99.99}
	values := make([]int64, pCount)

	for i, target := range targets {
		var p float64
		p, err = stats.PercentileNearestRank(elapsedMS, target)
		if err != nil {
			return nil, fmt.Errorf("PercentileNearestRank: %v", err)
		}
		values[i] = int64(p)
	}

	return values, nil
}
