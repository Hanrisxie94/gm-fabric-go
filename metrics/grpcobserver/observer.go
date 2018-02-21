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
// limitations under theAccumulate License.

package grpcobserver

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/deciphernow/gm-fabric-go/metrics/subject"
)

// Accumulate updates a stats struct from an event struct
// it returns true if the struct is complete
func Accumulate(entry APIStats, event subject.MetricsEvent) (APIStats, bool) {
	var end bool

	switch event.EventType {
	case "rpc.InHeader":
		entry.Key = event.Key
		entry.Transport = event.Transport
		entry.PrevRoute = event.PrevRoute
		entry.InWireLength += numericEventValue(event.Value)
	case "rpc.Begin":
		entry.BeginTime = event.Timestamp
	case "rpc.InPayload":
		entry.InWireLength += numericEventValue(event.Value)
	case "rpc.InTrailer":
		entry.InWireLength += numericEventValue(event.Value)
	case "rpc.OutPayload":
		entry.OutWireLength += numericEventValue(event.Value)
	case "rpc.OutTrailer":
		entry.OutWireLength += numericEventValue(event.Value)
	case "rpc.End":
		entry.EndTime = event.Timestamp
		entry.HTTPStatus = event.HTTPStatus
		if event.Value != nil {
			entry.Err = event.Value.(error)
		}
		end = true
	}

	return entry, end
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

// APIStats reports stats on an individual API call
type APIStats struct {
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

type keyEventsEntry struct {
	events            int64
	statusEvents      map[int]int64
	statusClassEvents map[string]int64
}

type cumulativeCounts struct {
	totalEvents     int64
	transportEvents map[subject.EventTransport]int64
	keyEvents       map[string]keyEventsEntry
}

// LatencyStatsGetter provides access to latency statistics
type LatencyStatsGetter interface {

	// GetLatencyStats returns a snapshot of the current latency statistics
	GetLatencyStats() (map[string]APIEndpointStats, error)
}

// GRPCObserver implements the Observer interface. Also supports HTTP handlers.
// Also implements the LatencyStatsGetter interface.
type GRPCObserver struct {
	sync.Mutex

	startTime time.Time

	active map[string]APIStats
	cache  *apiStatsCache

	cumulativeCounts
}

// New returns an entity that supports the Observer interface and which
// can register HTTP handler functions
func New(
	cacheSize int,
) *GRPCObserver {
	return &GRPCObserver{
		startTime:        time.Now().UTC(),
		active:           make(map[string]APIStats),
		cache:            newAPIStatsCache(cacheSize),
		cumulativeCounts: newCumulativeCounts(),
	}
}

func newCumulativeCounts() cumulativeCounts {
	return cumulativeCounts{
		transportEvents: make(map[subject.EventTransport]int64),
		keyEvents:       make(map[string]keyEventsEntry),
	}
}

func copyCumulativeCounts(inp cumulativeCounts) cumulativeCounts {
	outp := newCumulativeCounts()
	outp.totalEvents = inp.totalEvents
	for key, value := range inp.keyEvents {
		outp.keyEvents[key] = value
	}
	for key, value := range inp.transportEvents {
		outp.transportEvents[key] = value
	}

	return outp
}

// Observe implements the subject.Observer interface, an instance of
// the observer design pattern
func (obs *GRPCObserver) Observe(event subject.MetricsEvent) {
	if !strings.HasPrefix(event.EventType, "rpc.") {
		return
	}

	obs.Lock()
	defer obs.Unlock()

	entry := obs.active[event.RequestID]
	entry, end := Accumulate(entry, event)

	if end {
		obs.totalEvents++
		obs.transportEvents[entry.Transport]++
		keyEvents, ok := obs.keyEvents[entry.Key]
		if !ok {
			keyEvents.statusEvents = make(map[int]int64)
			keyEvents.statusClassEvents = make(map[string]int64)
		}
		keyEvents.events++
		if entry.Transport == subject.EventTransportHTTP ||
			entry.Transport == subject.EventTransportHTTPS {
			keyEvents.statusEvents[entry.HTTPStatus]++
			keyEvents.statusClassEvents[statusClass(entry.HTTPStatus)]++
		}
		obs.keyEvents[entry.Key] = keyEvents
		obs.cache.store(entry)
		delete(obs.active, event.RequestID)
	} else {
		obs.active[event.RequestID] = entry
	}
}

func numericEventValue(rawValue interface{}) int64 {
	result, _ := rawValue.(int64)
	return result
}

// statusClass returns a string of the form NXX for HTTP Status, e.g. 2XX
func statusClass(status int) string {
	return fmt.Sprintf("%dXX", status/100)
}
