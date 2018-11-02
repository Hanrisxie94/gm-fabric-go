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
	input  APIStatsEntry
	probes []probefunc
}

func TestAPIStats(t *testing.T) {
	const key = "xxx"
	testCases := []testEntry{
		testEntry{
			name:  "zero latency",
			input: APIStatsEntry{Key: key, BeginTime: time.Now(), EndTime: time.Now()},
			probes: []probefunc{
				func(ep APIEndpointStats) error {
					if ep.Sum != 0 {
						return errors.Errorf("Expected 0 latency: found %d", ep.Sum)
					}
					return nil
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, tc.name), func(t *testing.T) {
			stats := New(1)
			stats.Store(tc.input)
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
