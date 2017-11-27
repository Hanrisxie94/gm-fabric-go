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

package grpcobserver

import (
	"fmt"
	"testing"
)

func TestAPICache(t *testing.T) {
	const size = 3
	testEntries := make([]APIStats, 2*size)
	for i := 0; i < len(testEntries); i++ {
		testEntries[i] = APIStats{Key: fmt.Sprintf("%03d", i)}
	}
	testCases := []struct {
		storeIndex      int
		expectedSize    int
		expectedContent []int
	}{
		{
			storeIndex:      0,
			expectedSize:    1,
			expectedContent: []int{0},
		},
		{
			storeIndex:      1,
			expectedSize:    2,
			expectedContent: []int{0, 1},
		},
		{
			storeIndex:      2,
			expectedSize:    3,
			expectedContent: []int{0, 1, 2},
		},
		{
			storeIndex:      3,
			expectedSize:    3,
			expectedContent: []int{1, 2, 3},
		},
		{
			storeIndex:      4,
			expectedSize:    3,
			expectedContent: []int{2, 3, 4},
		},
		{
			storeIndex:      5,
			expectedSize:    3,
			expectedContent: []int{3, 4, 5},
		},
	}

	c := newAPIStatsCache(size)
	if c.size() != 0 {
		t.Fatalf("size: expected 0 found %d", c.size())
	}
	content := getCacheContent(c)
	if len(content) != 0 {
		t.Fatalf("getCacheContent: expected 0, found %q", content)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			c.store(testEntries[tc.storeIndex])
			if c.size() != tc.expectedSize {
				t.Fatalf("size: expected %d found %d", tc.expectedSize, c.size())
			}
			content = getCacheContent(c)
			if len(content) != len(tc.expectedContent) {
				t.Fatalf("getCacheContent: expected %v, found %q",
					tc.expectedContent, content)
			}
			for j, k := range tc.expectedContent {
				if content[j].Key != testEntries[k].Key {
					t.Fatalf("getCacheContent: expected %s, found %s",
						testEntries[k].Key, content[i].Key)
				}
			}
		})
	}

}

func getCacheContent(c *apiStatsCache) []APIStats {
	var content []APIStats
	for entry := range c.traverse() {
		content = append(content, entry)
	}

	return content
}
