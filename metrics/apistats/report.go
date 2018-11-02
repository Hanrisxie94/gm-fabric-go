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

	"github.com/deciphernow/gm-fabric-go/metrics/flatjson"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
	"github.com/pkg/errors"
)

// Report implements the Reporter interface it is called by the metrics server
func (st *APIStats) Report(jWriter *flatjson.Writer) error {
	var err error
	var summary APIStatsSummary
	var counts CumulativeCounts

	if summary.APIStats, err = st.GetEndpointStats(); err != nil {
		return errors.Wrap(err, "st.GetEndpointStats()")
	}
	counts = st.GetCumulativeCounts()

	err = jWriter.Write(fmt.Sprintf("%s/%s", "Total", "requests"), counts.TotalEvents)
	if err != nil {
		return errors.Wrap(err, "jWriter.Write Total requests")
	}

	if err = st.writeTransport(jWriter, counts); err != nil {
		return errors.Wrap(err, "writeTransport")
	}

	allEvents := accumulateAllEvents(summary, counts)

	for path, value := range summary.APIStats {
		var keyEvents KeyEventsEntry
		if path == "all" {
			keyEvents = allEvents
		} else {
			keyEvents = counts.KeyEvents[path]
		}
		err = writeValue(path, value, keyEvents, jWriter)
		if err != nil {
			return errors.Wrap(err, "writeValue")
		}
	}

	return nil
}

func (st *APIStats) writeTransport(
	jWriter *flatjson.Writer,
	counts CumulativeCounts,
) error {
	transportLabels := map[subject.EventTransport]string{
		subject.EventTransportHTTP:       "HTTP",
		subject.EventTransportHTTPS:      "HTTPS",
		subject.EventTransportRPC:        "RPC",
		subject.EventTransportRPCWithTLS: "RPC_TLS",
	}

	for _, transport := range []subject.EventTransport{
		subject.EventTransportHTTP,
		subject.EventTransportHTTPS,
		subject.EventTransportRPC,
		subject.EventTransportRPCWithTLS,
	} {
		err := jWriter.Write(
			fmt.Sprintf("%s/%s", transportLabels[transport], "requests"),
			counts.TransportEvents[transport],
		)
		if err != nil {
			return errors.Wrapf(err, "jWriter.Write %s requests",
				transportLabels[transport])
		}
	}

	return nil
}

func accumulateAllEvents(
	summary APIStatsSummary,
	counts CumulativeCounts,
) KeyEventsEntry {
	allEvents := KeyEventsEntry{
		StatusEvents:      make(map[int]int64),
		StatusClassEvents: make(map[string]int64),
	}

	for path := range summary.APIStats {
		if path != "all" {
			keyEvents := counts.KeyEvents[path]
			allEvents.Events += keyEvents.Events
			for key, value := range keyEvents.StatusEvents {
				allEvents.StatusEvents[key] += value
			}
			for key, value := range keyEvents.StatusClassEvents {
				allEvents.StatusClassEvents[key] += value
			}
		}
	}

	return allEvents
}

func writeValue(
	path string,
	value APIEndpointStats,
	keyEvents KeyEventsEntry,
	jWriter *flatjson.Writer,
) error {
	err := jWriter.Write(fmt.Sprintf("%s/%s", path, "requests"), keyEvents.Events)
	if err != nil {
		return errors.Wrapf(err, "jWriter.Write %s requests", path)
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
		return errors.Wrapf(err, "jWriter.Write %s routes", path)
	}

	for stat, statValue := range keyEvents.StatusEvents {
		err = jWriter.Write(fmt.Sprintf("%s/status/%d", path, stat), statValue)
		if err != nil {
			return errors.Wrapf(err, "jWriter.Write %s status", path)
		}
	}

	for statClass, statClassValue := range keyEvents.StatusClassEvents {
		err = jWriter.Write(fmt.Sprintf("%s/status/%s", path, statClass), statClassValue)
		if err != nil {
			return errors.Wrapf(err, "jWriter.Write %s statClass", path)
		}
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
			return errors.Wrapf(err, "jWriter.Write %s/%s", path, x.label)
		}
	}

	return nil
}
