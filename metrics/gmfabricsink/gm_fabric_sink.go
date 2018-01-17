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
func (s *sinkStruct) SetGauge(key []string, val float32) {
	s.eventChan <- subject.MetricsEvent{
		EventType: "go-metrics.SetGauge",
		Key:       joinKey(key),
		Timestamp: time.Now().UTC(),
		Value:     val,
	}
}

// SetGaugeWithLabels should retain the last value it is set to
func (s *sinkStruct) SetGaugeWithLabels(
	key []string,
	val float32,
	labels []gometrics.Label,
) {
	s.eventChan <- subject.MetricsEvent{
		EventType: "go-metrics.SetGaugeWithLabels",
		Key:       joinKey(key),
		Timestamp: time.Now().UTC(),
		Value:     val,
		Tags:      labels2tags(labels),
	}
}

// EmitKey should emit a Key/Value pair for each call
func (s *sinkStruct) EmitKey(key []string, val float32) {
	s.eventChan <- subject.MetricsEvent{
		EventType: "go-metrics.EmitKey",
		Key:       joinKey(key),
		Timestamp: time.Now().UTC(),
		Value:     val,
	}
}

// IncrCounter should accumulate a value in a counter
func (s *sinkStruct) IncrCounter(key []string, val float32) {
	s.eventChan <- subject.MetricsEvent{
		EventType: "go-metrics.IncrCounter",
		Key:       joinKey(key),
		Timestamp: time.Now().UTC(),
		Value:     val,
	}
}

// IncrCounterWithLabels should accumulate a value in a counter
func (s *sinkStruct) IncrCounterWithLabels(
	key []string,
	val float32,
	labels []gometrics.Label,
) {
	s.eventChan <- subject.MetricsEvent{
		EventType: "go-metrics.IncrCounterWithLabels",
		Key:       joinKey(key),
		Timestamp: time.Now().UTC(),
		Value:     val,
		Tags:      labels2tags(labels),
	}
}

// AddSample for timing information, where quantiles are used
func (s *sinkStruct) AddSample(key []string, val float32) {
	s.eventChan <- subject.MetricsEvent{
		EventType: "go-metrics.AddSample",
		Key:       joinKey(key),
		Timestamp: time.Now().UTC(),
		Value:     val,
	}
}

// AddSampleWithLabels for timing information, where quantiles are used
func (s *sinkStruct) AddSampleWithLabels(
	key []string,
	val float32,
	labels []gometrics.Label,
) {
	s.eventChan <- subject.MetricsEvent{
		EventType: "go-metrics.AddSampleWithLabels",
		Key:       joinKey(key),
		Timestamp: time.Now().UTC(),
		Value:     val,
		Tags:      labels2tags(labels),
	}
}

func joinKey(key []string) string {
	// we assume the key is of the form [service, host, ...]
	// if so, we leave out service and host
	if len(key) > 2 {
		return strings.Join(key[2:], "/")
	}
	return strings.Join(key, "/")
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
