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
	"strings"
	"sync"
	"time"

	"github.com/deciphernow/gm-fabric-go/metrics/apistats"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
)

// Accumulate updates a stats struct from an event struct
// it returns true if the struct is complete
func Accumulate(entry apistats.APIStatsEntry, event subject.MetricsEvent) (apistats.APIStatsEntry, bool) {
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

// GRPCObserver implements the Observer interface. Also supports HTTP handlers.
// Also implements the LatencyStatsGetter interface.
type GRPCObserver struct {
	sync.Mutex

	startTime time.Time

	active   map[string]apistats.APIStatsEntry
	apiStats *apistats.APIStats
}

// New returns an entity that supports the Observer interface and which
// can register HTTP handler functions
func New(
	cacheSize int,
) *GRPCObserver {
	return &GRPCObserver{
		active:   make(map[string]apistats.APIStatsEntry),
		apiStats: apistats.New(cacheSize),
	}
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
		obs.apiStats.Store(entry)
		delete(obs.active, event.RequestID)
	} else {
		obs.active[event.RequestID] = entry
	}
}

func numericEventValue(rawValue interface{}) int64 {
	result, _ := rawValue.(int64)
	return result
}
