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

// apiCache is a circular queue of APIStats
type apiStatsCache struct {
	start      int
	end        int
	activeSize int
	cache      []APIStats
}

func newAPIStatsCache(size int) *apiStatsCache {
	var ec apiStatsCache
	ec.cache = make([]APIStats, size)
	return &ec
}

func (ec *apiStatsCache) size() int {
	return ec.activeSize
}

func (ec *apiStatsCache) store(entry APIStats) {
	if ec.activeSize < len(ec.cache) {
		ec.activeSize++
		ec.end = ec.activeSize - 1
		ec.cache[ec.end] = entry
	} else {
		ec.end = (ec.end + 1) % len(ec.cache)
		ec.cache[ec.end] = entry
		ec.start = (ec.end + 1) % len(ec.cache)
	}
}

// traverse iterates through the cache, pushing entries into the channel
// they should return in the order they were stored (in order by BeginTime)
func (ec *apiStatsCache) traverse() <-chan APIStats {
	ch := make(chan APIStats)
	go func() {
		if ec.activeSize == 0 {
			close(ch)
			return
		}
		for i := ec.start; ; i = (i + 1) % len(ec.cache) {
			ch <- ec.cache[i]
			if i == ec.end {
				close(ch)
				return
			}
		}
	}()

	return ch
}
