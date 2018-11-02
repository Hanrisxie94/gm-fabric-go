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

// APIStatsCache is a circular queue of APIStats
type APIStatsCache struct {
	start      int
	end        int
	activeSize int
	cache      []APIStatsEntry
}

func NewAPIStatsCache(size int) *APIStatsCache {
	var ec APIStatsCache
	ec.cache = make([]APIStatsEntry, size)
	return &ec
}

// Size returns a count of the active entries in the cache
func (ec *APIStatsCache) Size() int {
	return ec.activeSize
}

// Store appends a new entry to the cache, writing over the oldest if necessary
func (ec *APIStatsCache) Store(entry APIStatsEntry) {
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

// Traverse iterates through the cache, pushing entries into the channel
// they should return in the order they were stored (in order by BeginTime)
func (ec *APIStatsCache) Traverse() <-chan APIStatsEntry {
	ch := make(chan APIStatsEntry)
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
