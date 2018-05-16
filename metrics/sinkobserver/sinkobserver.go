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

package sinkobserver

import (
	"fmt"
	"strings"
	"sync"
	"time"

	gometrics "github.com/armon/go-metrics"

	"github.com/deciphernow/gm-fabric-go/metrics/grpcobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/memvalues"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
)

type activeEntry struct {
	stats  grpcobserver.APIStats
	tagMap map[string]string
}

type sinkObs struct {
	sync.Mutex
	active map[string]activeEntry
	sink   gometrics.MetricSink
}

// New return an observer that feeds the go-metrics sink
func New(
	sink gometrics.MetricSink,
	reportInterval time.Duration,
) subject.Observer {
	obs := sinkObs{
		sink:   sink,
		active: make(map[string]activeEntry),
	}
	go obs.reportMemory(reportInterval)
	return &obs
}

// Observe implements the Observer pattern
func (so *sinkObs) Observe(event subject.MetricsEvent) {
	so.Lock()
	defer so.Unlock()

	entry := so.active[event.RequestID]

	// TODO: we are using the APIStats object here because that's
	// what the system was originally designed to collect. We should
	// probably either move it to it's own package, or use different handlers
	entryStats, end := grpcobserver.Accumulate(entry.stats, event)
	entry.stats = entryStats
	if entry.tagMap == nil {
		entry.tagMap = make(map[string]string)
	}
	updateTagMap(entry.tagMap, event)

	if end {
		key := []string{
			entry.tagMap["service"],
			entry.tagMap["host"],
			splitEntryKey(entry.stats.Key),
		}
		elapsed := entry.stats.EndTime.Sub(entry.stats.BeginTime)
		so.sink.IncrCounter(
			append(key, "in_throughput"),
			float32(entry.stats.InWireLength),
		)
		so.sink.IncrCounter(
			append(key, "out_throughput"),
			float32(entry.stats.OutWireLength),
		)
		if entry.stats.Err != nil {
			so.sink.IncrCounter(append(key, "errors"), 1)
		}
		so.sink.AddSample(
			append(key, "latency"),
			duration2ms(elapsed),
		)
		delete(so.active, event.RequestID)
	} else {
		so.active[event.RequestID] = entry
	}
}

func (so *sinkObs) reportMemory(reportInterval time.Duration) {
	tickChan := time.Tick(reportInterval)
	for {
		<-tickChan
		memValues, _ := memvalues.GetMemValues()
		so.Lock()
		so.sink.AddSample(
			[]string{"memory", "system", "available"},
			float32(memValues.SystemMemoryAvailable),
		)
		so.sink.AddSample(
			[]string{"memory", "system", "used"},
			float32(memValues.SystemMemoryUsed),
		)
		so.sink.AddSample(
			[]string{"memory", "system", "used-percent"},
			float32(memValues.SystemMemoryUsedPercent),
		)
		so.sink.AddSample(
			[]string{"memory", "process", "used"},
			float32(memValues.ProcessMemoryUsed),
		)
		so.Unlock()
	}
}

func duration2ms(d time.Duration) float32 {
	const nsPerMs = 1000000

	return float32(d.Nanoseconds() / nsPerMs)
}

func updateTagMap(tagMap map[string]string, event subject.MetricsEvent) {
	for _, tag := range event.Tags {
		name, value := subject.SplitTag(tag)
		if value != "" {
			_, ok := tagMap[name]
			if !ok {
				tagMap[name] = value
			}
		}
	}
}

// splitEntryKey cleans up the key by removing slashes
// We have to handle HTTP routes separately because they include http method,
// GET, POST, etc
func splitEntryKey(rawKey string) string {
	splitKey := strings.Split(rawKey, "/")

	if len(splitKey) == 0 {
		return rawKey
	}

	// take a key of the form route/movie/GET and return movie(GET)
	if len(splitKey) > 2 && splitKey[0] == "route" {
		uriKey := splitKey[len(splitKey)-2]
		httpMethod := splitKey[len(splitKey)-1]

		return fmt.Sprintf("%s(%s)", uriKey, httpMethod)
	}

	// return the last element
	return splitKey[len(splitKey)-1]
}
