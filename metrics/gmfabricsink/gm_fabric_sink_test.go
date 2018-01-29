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
	"testing"

	gometrics "github.com/armon/go-metrics"

	"github.com/deciphernow/gm-fabric-go/metrics/subject"
)

func TestSink(t *testing.T) {
	tests := []struct {
		eventType string
		key       []string
		val       float32
	}{
		{"go-metrics.SetGauge", []string{"aaa"}, 1},
		{"go-metrics.SetGaugeWithLabels", []string{"bbb"}, 2},
		{"go-metrics.EmitKey", []string{"ccc"}, 3},
		{"go-metrics.IncrCounter", []string{"ddd"}, 4},
		{"go-metrics.IncrCounterWithLabels", []string{"eee"}, 5},
		{"go-metrics.AddSample", []string{"fff"}, 6},
		{"go-metrics.AddSampleWithLabels", []string{"ggg"}, 7},
	}

	eventChan := make(chan subject.MetricsEvent, len(tests))
	sink := New(eventChan)

	for _, test := range tests {
		switch test.eventType {
		case "go-metrics.SetGauge":
			sink.SetGauge(test.key, test.val)
		case "go-metrics.SetGaugeWithLabels":
			sink.SetGaugeWithLabels(test.key, test.val, nil)
		case "go-metrics.EmitKey":
			sink.EmitKey(test.key, test.val)
		case "go-metrics.IncrCounter":
			sink.IncrCounter(test.key, test.val)
		case "go-metrics.IncrCounterWithLabels":
			sink.IncrCounterWithLabels(test.key, test.val, nil)
		case "go-metrics.AddSample":
			sink.AddSample(test.key, test.val)
		case "go-metrics.AddSampleWithLabels":
			sink.AddSampleWithLabels(test.key, test.val, nil)
		}
	}

	close(eventChan)

	var i int
	for event := range eventChan {

		test := tests[i]

		if event.EventType != test.eventType {
			t.Fatalf("event type mismatch %s != %s",
				event.EventType, test.eventType)
		}
		if event.Key != test.key[0] {
			t.Fatalf("key mismatch %s != %s", event.Key, test.key[0])
		}
		if event.Value != test.val {
			t.Fatalf("value mismatch %f != %f", event.Value, test.val)
		}

		i++
	}
}

func Test_joinKey(t *testing.T) {
	type args struct {
		key []string
	}
	tests := []struct {
		name         string
		args         args
		expectedKey  string
		expectedTags []string
	}{
		{"empty", args{nil}, "", nil},
		{"single", args{[]string{"aaa"}}, "aaa", nil},
		{"double", args{[]string{"aaa", "bbb"}}, "aaa/bbb", nil},
		{"triple", args{[]string{"aaa", "bbb", "ccc"}}, "ccc", []string{"service:aaa", "host:bbb"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, tags := joinKey(tt.args.key)
			if key != tt.expectedKey {
				t.Errorf("joinKey() key = %v, want %v", key, tt.expectedKey)
			}
			if tt.expectedTags == nil && tags != nil {
				t.Errorf("joinKey() tags = %v, want %v", tags, tt.expectedTags)
			}
			if tt.expectedTags != nil && tags == nil {
				t.Errorf("joinKey() tags = %v, want %v", tags, tt.expectedTags)
			}
			if len(tags) != len(tt.expectedTags) {
				t.Errorf("joinKey() tags = %v, want %v", tags, tt.expectedTags)
			}
			for i := 0; i < len(tt.expectedTags); i++ {
				if tags[i] != tt.expectedTags[i] {
					t.Errorf("joinKey() tags[%d] = %v, want %v", i, tags[i], tt.expectedTags[i])
				}
			}
		})
	}
}

func TestLabels2tags(t *testing.T) {

	for i, td := range []struct {
		labels   []gometrics.Label
		expected []string
	}{
		{
			labels:   nil,
			expected: nil,
		},
		{
			labels: []gometrics.Label{
				gometrics.Label{
					Name:  "",
					Value: "",
				},
			},
			expected: nil,
		},
		{
			labels: []gometrics.Label{
				gometrics.Label{
					Name:  "aaa",
					Value: "",
				},
			},
			expected: []string{"aaa"},
		},
		{
			labels: []gometrics.Label{
				gometrics.Label{
					Name:  "aaa",
					Value: "bbb",
				},
			},
			expected: []string{"aaa:bbb"},
		},
	} {
		tags := labels2tags(td.labels)
		if len(tags) != len(td.expected) {
			t.Fatalf("#%d: size mismatch %v != %v", i+1, tags, td.expected)
		}
		for j := 0; j < len(tags); j++ {
			if tags[j] != td.expected[j] {
				t.Fatalf("#%d: %v != %v", i+1, tags, td.expected)
			}
		}
	}

}
