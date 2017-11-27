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
	"sync"
	"time"

	gometrics "github.com/armon/go-metrics"

	"github.com/deciphernow/gm-fabric-go/metrics/grpcobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/memvalues"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
)

type sinkObs struct {
	sync.Mutex
	active map[string]grpcobserver.APIStats
	sink   gometrics.MetricSink
}

// New return an observer that feeds the go-metrics sink
func New(
	sink gometrics.MetricSink,
	reportInterval time.Duration,
) subject.Observer {
	obs := sinkObs{
		sink:   sink,
		active: make(map[string]grpcobserver.APIStats),
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
	entry, end := grpcobserver.Accumulate(entry, event)

	if end {
		elapsed := entry.EndTime.Sub(entry.BeginTime)
		so.sink.IncrCounter(
			[]string{"in", entry.Key},
			float32(entry.InWireLength),
		)
		so.sink.IncrCounter(
			[]string{"out", entry.Key},
			float32(entry.OutWireLength),
		)
		if entry.Err != nil {
			so.sink.IncrCounter([]string{"errors", entry.Key}, 1)
		}
		so.sink.AddSample(
			[]string{"latency", entry.Key},
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
