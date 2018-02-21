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

package gmfabricsink

import (
	"strings"
	"time"

	gometrics "github.com/armon/go-metrics"

	"github.com/deciphernow/gm-fabric-go/metrics/subject"
)

// New returns an entity that implements the go-metrics MetricSink interface.
// It sends MetricsEvent objects to the metrics server
func New(eventChan chan<- subject.MetricsEvent) gometrics.MetricSink {
	return &sinkStruct{eventChan: eventChan}
}

type sinkStruct struct {
	eventChan chan<- subject.MetricsEvent
}

// SetGauge should retain the last value it is set to
func (s *sinkStruct) SetGauge(keys []string, val float32) {
	s.emit("go-metrics.SetGauge", keys, nil, val)
}

// SetGaugeWithLabels should retain the last value it is set to
func (s *sinkStruct) SetGaugeWithLabels(keys []string, val float32, labels []gometrics.Label) {
	s.emit("go-metrics.SetGaugeWithLabels", keys, labels, val)
}

// EmitKey should emit a Key/Value pair for each call
func (s *sinkStruct) EmitKey(keys []string, val float32) {
	s.emit("go-metrics.EmitKey", keys, nil, val)
}

// IncrCounter should accumulate a value in a counter
func (s *sinkStruct) IncrCounter(keys []string, val float32) {
	s.emit("go-metrics.IncrCounter", keys, nil, val)
}

// IncrCounterWithLabels should accumulate a value in a counter
func (s *sinkStruct) IncrCounterWithLabels(keys []string, val float32, labels []gometrics.Label) {
	s.emit("go-metrics.IncrCounterWithLabels", keys, labels, val)
}

// AddSample for timing information, where quantiles are used
func (s *sinkStruct) AddSample(keys []string, val float32) {
	s.emit("go-metrics.AddSample", keys, nil, val)
}

// AddSampleWithLabels for timing information, where quantiles are used
func (s *sinkStruct) AddSampleWithLabels(keys []string, val float32, labels []gometrics.Label) {
	s.emit("go-metrics.AddSampleWithLabels", keys, labels, val)
}

func (sink *sinkStruct) emit(eventType string, keys []string, labels []gometrics.Label, value float32) {
	key, tags := joinKey(keys)
	if labels != nil {
		tags = append(tags, labels2tags(labels)...)
	}
	sink.eventChan <- subject.MetricsEvent {
		EventType: eventType,
		Key:       key,
		Timestamp: time.Now().UTC(),
		Value:     value,
		Tags:      tags,
	}
}

func joinKey(key []string) (string, []string) {
	// we assume the key is of the form [service, host, ...]
	// if so, we leave out service and host
	if len(key) > 2 {
		serviceTag := subject.JoinTag("service", key[0])
		hostTag := subject.JoinTag("host", key[1])
		return strings.Join(key[2:], "/"), []string{serviceTag, hostTag}
	}
	return strings.Join(key, "/"), nil
}

func labels2tags(labels []gometrics.Label) []string {
	var tags []string

	for _, label := range labels {
		tag := subject.JoinTag(label.Name, label.Value)
		if len(tag) > 0 {
			tags = append(tags, tag)
		}
	}

	return tags
}
