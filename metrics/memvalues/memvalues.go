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

package memvalues

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/mem"
)

// MemValues holds memory values for reporting
type MemValues struct {
	SystemMemoryAvailable   uint64
	SystemMemoryUsed        uint64
	SystemMemoryUsedPercent float64
	ProcessMemoryUsed       uint64
}

// GetMemValues returns current memory values
func GetMemValues() (MemValues, error) {
	var vmStats *mem.VirtualMemoryStat
	var memStats runtime.MemStats
	var err error

	vmStats, err = mem.VirtualMemory()
	if err != nil {
		return MemValues{}, fmt.Errorf("mem.VirtualMemory() failed: %v", err)
	}

	runtime.ReadMemStats(&memStats)

	return MemValues{
		SystemMemoryAvailable:   vmStats.Available,
		SystemMemoryUsed:        vmStats.Used,
		SystemMemoryUsedPercent: vmStats.UsedPercent,
		ProcessMemoryUsed:       memStats.Sys,
	}, nil
}
