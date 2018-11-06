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
	"testing"
	"time"

	"github.com/pkg/errors"
)

type probefunc func(APIEndpointStats) error

type testEntry struct {
	name   string
	input  []APIStatsEntry
	probes []probefunc
}

func TestAPIStats(t *testing.T) {
	const key = "xxx"

	// get this time once, so we don't pick up little differences
	currentTime := time.Now()
	beginTime := currentTime
	var expectedLatency int64 = 42 // ms
	requestTime := beginTime.Add(time.Millisecond * time.Duration(expectedLatency))
	var inCaptureSeconds int64 = 2
	inCaptureTime := requestTime.Add(time.Second * time.Duration(inCaptureSeconds))
	var inWireLength int64 = 4096
	inExpectedThroughput := inWireLength / inCaptureSeconds
	responseTime := requestTime.Add(time.Millisecond * time.Duration(33))
	var outCaptureSeconds int64 = 3
	outCaptureTime := responseTime.Add(time.Second * time.Duration(outCaptureSeconds))
	var outWireLength int64 = 999999
	outExpectedThroughput := outWireLength / outCaptureSeconds

	testCases := []testEntry{
		// handle the null case
		// We expect zero latency and zero throughput
		testEntry{
			name:  "zero data",
			input: []APIStatsEntry{APIStatsEntry{Key: key}},
			probes: []probefunc{
				func(ep APIEndpointStats) error {
					if ep.Sum != 0 {
						return errors.Errorf("Expected 0 latency: found %d", ep.Sum)
					}
					return nil
				},
				func(ep APIEndpointStats) error {
					if ep.InThroughput != 0 {
						return errors.Errorf("Expected 0 in throughput: found %d", ep.InThroughput)
					}
					if ep.OutThroughput != 0 {
						return errors.Errorf("Expected 0 out throughput: found %d", ep.OutThroughput)
					}
					return nil
				},
			},
		},
		// test simple latency
		// BeginTime = now,
		// RequestTime = BeginTime + 42ms
		// expecting latency of 42ms,
		testEntry{
			name: "simple latency",
			input: []APIStatsEntry{
				APIStatsEntry{
					Key:         key,
					BeginTime:   beginTime,
					RequestTime: requestTime,
				},
			},
			probes: []probefunc{
				func(ep APIEndpointStats) error {
					if ep.Sum != expectedLatency {
						return errors.Errorf("Expected latency %d: found %d",
							expectedLatency, ep.Sum)
					}
					return nil
				},
			},
		},
		// test simple incoming throughput
		// BeginTime = now,
		// RequestTime = BeginTime + 42ms
		// InCaptureTime = RequestTime + 2s
		// InWireLength = 4096 bytes
		// expecting latency of 42ms,
		// In throughput of (4096 / 2) bytes/sec = 2048
		testEntry{
			name: "simple incoming throughput",
			input: []APIStatsEntry{
				APIStatsEntry{
					Key:           key,
					BeginTime:     beginTime,
					RequestTime:   requestTime,
					InWireLength:  inWireLength,
					InCaptureTime: inCaptureTime,
				},
			},
			probes: []probefunc{
				func(ep APIEndpointStats) error {
					if ep.Sum != expectedLatency {
						return errors.Errorf("Expected latency %d: found %d",
							expectedLatency, ep.Sum)
					}
					return nil
				},
				func(ep APIEndpointStats) error {
					if ep.InThroughput != inExpectedThroughput {
						return errors.Errorf("Expected %d in throughput: found %d",
							inExpectedThroughput, ep.InThroughput)
					}
					if ep.OutThroughput != 0 {
						return errors.Errorf("Expected 0 out throughput: found %d", ep.OutThroughput)
					}
					return nil
				},
			},
		},
		// test simple outgoing throughput
		// BeginTime = now,
		// RequestTime = BeginTime + 42ms
		// InCaptureTime = RequestTime + 2s
		// InWireLength = 4096 bytes
		// ResponseTime = RequestTime + 33ms
		// OutCaptureTime = ResponseTime + 3s
		// OutWireLength = 999999 bytes
		// expecting latency of 42ms,
		// OutIn throughput of (999999 / 3) bytes/sec = 333333
		testEntry{
			name: "simple outgoing throughput",
			input: []APIStatsEntry{
				APIStatsEntry{
					Key:            key,
					BeginTime:      beginTime,
					RequestTime:    requestTime,
					ResponseTime:   responseTime,
					OutWireLength:  outWireLength,
					OutCaptureTime: outCaptureTime,
				},
			},
			probes: []probefunc{
				func(ep APIEndpointStats) error {
					if ep.Sum != expectedLatency {
						return errors.Errorf("Expected latency %d: found %d",
							expectedLatency, ep.Sum)
					}
					return nil
				},
				func(ep APIEndpointStats) error {
					if ep.OutThroughput != outExpectedThroughput {
						return errors.Errorf("Expected %d out throughput: found %d",
							outExpectedThroughput, ep.OutThroughput)
					}
					if ep.InThroughput != 0 {
						return errors.Errorf("Expected 0 in throughput: found %d", ep.InThroughput)
					}
					return nil
				},
			},
		},
		// test simple bidirectional throughput
		// BeginTime = now,
		// RequestTime = BeginTime + 42ms
		// ResponseTime = RequestTime + 33ms
		// OutCaptureTime = ResponseTime + 3s
		// OutWireLength = 999999 bytes
		// expecting latency of 42ms,
		// OutIn throughput of (999999 / 3) bytes/sec = 333333
		testEntry{
			name: "simple bidirectional throughput",
			input: []APIStatsEntry{
				APIStatsEntry{
					Key:            key,
					BeginTime:      beginTime,
					RequestTime:    requestTime,
					InWireLength:   inWireLength,
					InCaptureTime:  inCaptureTime,
					ResponseTime:   responseTime,
					OutWireLength:  outWireLength,
					OutCaptureTime: outCaptureTime,
				},
			},
			probes: []probefunc{
				func(ep APIEndpointStats) error {
					if ep.Sum != expectedLatency {
						return errors.Errorf("Expected latency %d: found %d",
							expectedLatency, ep.Sum)
					}
					return nil
				},
				func(ep APIEndpointStats) error {
					if ep.InThroughput != inExpectedThroughput {
						return errors.Errorf("Expected %d in throughput: found %d",
							inExpectedThroughput, ep.InThroughput)
					}
					if ep.OutThroughput != outExpectedThroughput {
						return errors.Errorf("Expected %d out throughput: found %d",
							outExpectedThroughput, ep.OutThroughput)
					}
					return nil
				},
			},
		},
		// test multiple bidirectional throughput
		// We should get the same results for multiple transactions
		// BeginTime = now,
		// RequestTime = BeginTime + 42ms
		// ResponseTime = RequestTime + 33ms
		// OutCaptureTime = ResponseTime + 3s
		// OutWireLength = 999999 bytes
		// expecting latency of 42ms,
		// OutIn throughput of (999999 / 3) bytes/sec = 333333
		testEntry{
			name: "simple bidirectional throughput",
			input: []APIStatsEntry{
				APIStatsEntry{
					Key:            key,
					BeginTime:      beginTime,
					RequestTime:    requestTime,
					InWireLength:   inWireLength,
					InCaptureTime:  inCaptureTime,
					ResponseTime:   responseTime,
					OutWireLength:  outWireLength,
					OutCaptureTime: outCaptureTime,
				},
				APIStatsEntry{
					Key:            key,
					BeginTime:      beginTime,
					RequestTime:    requestTime,
					InWireLength:   inWireLength,
					InCaptureTime:  inCaptureTime,
					ResponseTime:   responseTime,
					OutWireLength:  outWireLength,
					OutCaptureTime: outCaptureTime,
				},
				APIStatsEntry{
					Key:            key,
					BeginTime:      beginTime,
					RequestTime:    requestTime,
					InWireLength:   inWireLength,
					InCaptureTime:  inCaptureTime,
					ResponseTime:   responseTime,
					OutWireLength:  outWireLength,
					OutCaptureTime: outCaptureTime,
				},
				APIStatsEntry{
					Key:            key,
					BeginTime:      beginTime,
					RequestTime:    requestTime,
					InWireLength:   inWireLength,
					InCaptureTime:  inCaptureTime,
					ResponseTime:   responseTime,
					OutWireLength:  outWireLength,
					OutCaptureTime: outCaptureTime,
				},
			},
			probes: []probefunc{
				func(ep APIEndpointStats) error {
					if int64(ep.Avg) != expectedLatency {
						return errors.Errorf("Expected latency %d: found %d",
							expectedLatency, int64(ep.Avg))
					}
					return nil
				},
				func(ep APIEndpointStats) error {
					if ep.InThroughput != inExpectedThroughput {
						return errors.Errorf("Expected %d in throughput: found %d",
							inExpectedThroughput, ep.InThroughput)
					}
					if ep.OutThroughput != outExpectedThroughput {
						return errors.Errorf("Expected %d out throughput: found %d",
							outExpectedThroughput, ep.OutThroughput)
					}
					return nil
				},
			},
		},
		// test zero response time
		// this shouldn't happen, but it probably will
		// ResponseTime doesn't get set, because we never send a response
		// BeginTime = now,
		// RequestTime = 0
		// expecting latency of 42ms, throughput of zero
		testEntry{
			name: "zero response time",
			input: []APIStatsEntry{
				APIStatsEntry{
					Key:         key,
					BeginTime:   currentTime,
					RequestTime: currentTime.Add(time.Millisecond * time.Duration(expectedLatency)),
				},
			},
			probes: []probefunc{
				func(ep APIEndpointStats) error {
					if int64(ep.Avg) != expectedLatency {
						return errors.Errorf("Expected latency %d: found %d",
							expectedLatency, int64(ep.Avg))
					}
					return nil
				},
				func(ep APIEndpointStats) error {
					if ep.InThroughput != 0 {
						return errors.Errorf("Expected 0 in throughput: found %d", ep.InThroughput)
					}
					if ep.OutThroughput != 0 {
						return errors.Errorf("Expected 0 out throughput: found %d", ep.OutThroughput)
					}
					return nil
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, tc.name), func(t *testing.T) {
			stats := New(4096)
			for _, event := range tc.input {
				stats.Store(event)
			}
			result, err := stats.GetEndpointStats()
			if err != nil {
				t.Fatalf("stats.GetEndpointStats failed: %s", err)
			}
			entry, ok := result[key]
			if !ok {
				t.Fatalf("no entry for %s: %v", key, result)
			}
			for n, f := range tc.probes {
				err = f(entry)
				if err != nil {
					t.Fatalf("probe %d: %s", n, err)
				}
			}
		})
	}
}
