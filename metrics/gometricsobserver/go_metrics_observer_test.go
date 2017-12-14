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
	"bytes"
	"encoding/json"
	"testing"

	"github.com/deciphernow/gm-fabric-go/metrics/flatjson"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
)

func TestGoMetricsObserver(t *testing.T) {
	var err error

	tests := []subject.MetricsEvent{
		{
			EventType: "go-metrics.SetGauge",
			Key:       "aaa",
			Value:     float32(1),
		},
		{
			EventType: "go-metrics.SetGaugeWithLabels",
			Key:       "bbb",
			Value:     float32(2),
		},
		{
			EventType: "go-metrics.EmitKey",
			Key:       "ccc",
			Value:     float32(3),
		},
		{
			EventType: "go-metrics.IncrCounter",
			Key:       "ddd",
			Value:     float32(4),
		},
		{
			EventType: "go-metrics.IncrCounterWithLabels",
			Key:       "ddd",
			Value:     float32(5),
		},
		{
			EventType: "go-metrics.AddSample",
			Key:       "eee",
			Value:     float32(6),
		},
		{
			EventType: "go-metrics.AddSampleWithLabels",
			Key:       "fff",
			Value:     float32(7),
		},
	}
	observer := New()
	for _, test := range tests {
		observer.Observe(test)
	}
	if observer.GaugeMap["aaa"].Value != 1.0 {
		t.Fatalf("GaugeMap error: %f; expecting %f",
			observer.GaugeMap["aaa"].Value, 1.0)
	}
	if observer.GaugeMap["bbb"].Value != 2.0 {
		t.Fatalf("GaugeMap error: %f; expecting %f",
			observer.GaugeMap["bbb"].Value, 2.0)
	}
	if observer.EmitKeyMap["ccc"].Value != 3.0 {
		t.Fatalf("EmitKeyMap error: %f; expecting %f",
			observer.EmitKeyMap["ccc"].Value, 3.0)
	}
	if observer.CounterMap["ddd"].Value != 9.0 {
		t.Fatalf("CounterMap error: %f; expecting %f",
			observer.CounterMap["ddd"].Value, 9.0)
	}
	if observer.SampleMap["eee"].Value != 6.0 {
		t.Fatalf("SampleMap error: %f; expecting %f",
			observer.SampleMap["eee"].Value, 6.0)
	}
	if observer.SampleMap["fff"].Value != 7.0 {
		t.Fatalf("SampleMap error: %f; expecting %f",
			observer.SampleMap["fff"].Value, 7.0)
	}

	var buffer bytes.Buffer
	jr, err := flatjson.New(&buffer)
	if err != nil {
		t.Fatal(err)
	}
	if err = observer.Report(jr); err != nil {
		t.Fatal(err)
	}
	if err = jr.Flush(); err != nil {
		t.Fatal(err)
	}

	data := buffer.Bytes()
	var ts map[string]interface{}
	err = json.Unmarshal(data, &ts)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %s; %s", err, string(data))
	}

	if ts["go_metrics/aaa"].(float64) != 1.0 {
		t.Fatalf("ts.aaa expected %f %v", 1.0, string(data))
	}
	if ts["go_metrics/bbb"].(float64) != 2.0 {
		t.Fatalf("ts.bbb expected %f %v", 2.0, string(data))
	}
	if ts["go_metrics/ccc"].(float64) != 3.0 {
		t.Fatalf("ts.ccc expected %f %v", 3.0, string(data))
	}
	if ts["go_metrics/ddd"].(float64) != 9.0 {
		t.Fatalf("ts.ddd expected %f %v", 9.0, string(data))
	}
	if ts["go_metrics/eee"].(float64) != 6.0 {
		t.Fatalf("ts.eee expected %f %v", 6.0, string(data))
	}
	if ts["go_metrics/fff"].(float64) != 7.0 {
		t.Fatalf("ts.fff expected %f %v", 7.0, string(data))
	}

}
