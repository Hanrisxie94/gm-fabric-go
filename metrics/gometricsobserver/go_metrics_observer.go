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

package gometricsobserver

import (
	"strings"
	"sync"
	"time"

	"github.com/deciphernow/gm-fabric-go/metrics/subject"
)

// Gauge represents a metric that is set to the most recent value
type Gauge struct {
	Value     float32
	Timestamp time.Time
}

// EmitKey represents EmitKey
type EmitKey struct {
	Value     float32
	Timestamp time.Time
}

// Counter a counter that can be incremented
type Counter struct {
	Value     float32
	Timestamp time.Time
}

// Sample of ongoing data
type Sample struct {
	Value     float32
	Timestamp time.Time
}

// GoMetricsObserver implements the Observer interface and also
// supports http handlers
type GoMetricsObserver struct {
	sync.Mutex
	GaugeMap   map[string]Gauge
	EmitKeyMap map[string]EmitKey
	CounterMap map[string]Counter
	SampleMap  map[string]Sample
}

// New returns an entity that implements the Observer interface
func New() *GoMetricsObserver {
	return &GoMetricsObserver{
		GaugeMap:   make(map[string]Gauge),
		EmitKeyMap: make(map[string]EmitKey),
		CounterMap: make(map[string]Counter),
		SampleMap:  make(map[string]Sample),
	}
}

// Observe implements the Observer interface
func (g *GoMetricsObserver) Observe(event subject.MetricsEvent) {
	if !strings.HasPrefix(event.EventType, "go-metrics.") {
		return
	}

	g.Lock()
	defer g.Unlock()

	switch event.EventType {
	case "go-metrics.SetGauge":
		g.setGauge(event)
	case "go-metrics.SetGaugeWithLabels":
		g.setGauge(event)
	case "go-metrics.EmitKey":
		g.emitKey(event)
	case "go-metrics.IncrCounter":
		g.incrCounter(event)
	case "go-metrics.IncrCounterWithLabels":
		g.incrCounter(event)
	case "go-metrics.AddSample":
		g.addSample(event)
	case "go-metrics.AddSampleWithLabels":
		g.addSample(event)
	}
}

func (g *GoMetricsObserver) setGauge(event subject.MetricsEvent) {
	g.GaugeMap[event.Key] = Gauge{
		Value:     event.Value.(float32),
		Timestamp: event.Timestamp,
	}
}

func (g *GoMetricsObserver) emitKey(event subject.MetricsEvent) {
	g.EmitKeyMap[event.Key] = EmitKey{
		Value:     event.Value.(float32),
		Timestamp: event.Timestamp,
	}
}

func (g *GoMetricsObserver) incrCounter(event subject.MetricsEvent) {
	counter, _ := g.CounterMap[event.Key]
	counter.Value += event.Value.(float32)
	counter.Timestamp = event.Timestamp
	g.CounterMap[event.Key] = counter
}

func (g *GoMetricsObserver) addSample(event subject.MetricsEvent) {
	g.SampleMap[event.Key] = Sample{
		Value:     event.Value.(float32),
		Timestamp: event.Timestamp,
	}
}
