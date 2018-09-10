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

package prometheus

import (
	"runtime"
	"time"

	"github.com/deciphernow/gm-fabric-go/metrics/memvalues"
	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/cpu"
)

// ReportSystemMetrics periodically reports system metrics to Prometheus
func ReportSystemMetrics(
	collector Collector,
	interval time.Duration,
	logger zerolog.Logger,
) {
	ticker := time.NewTicker(interval)

METRICS_LOOP:
	for {
		var entry SystemMetricsEntry

		memValues, err := memvalues.GetMemValues()
		if err != nil {
			logger.Error().AnErr("memvalues.GetMemValues()", err).Msg("")
			continue METRICS_LOOP
		}

		var cpuPercent []float64
		cpuPercent, err = cpu.Percent(time.Second, false)
		if err != nil {
			logger.Error().AnErr("cpu.Percent", err).Msg("")
			continue METRICS_LOOP
		}

		entry.SystemCPUPercent = cpuPercent[0]
		entry.SystemCPUCores = float64(runtime.NumCPU())
		entry.SystemMemoryAvailable = float64(memValues.SystemMemoryAvailable)
		entry.SystemMemoryUsed = float64(memValues.SystemMemoryUsed)
		entry.SystemMemoryUsedPercent = memValues.SystemMemoryUsedPercent
		entry.ProcessMemoryUsed = float64(memValues.ProcessMemoryUsed)

		collector.CollectSystemMetrics(entry)

		<-ticker.C
	}
}
