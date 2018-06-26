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

package metricsserver

import (
	"runtime"
	"time"

	"github.com/pkg/errors"

	"github.com/deciphernow/gm-fabric-go/metrics/flatjson"
	"github.com/deciphernow/gm-fabric-go/metrics/memvalues"
	"github.com/shirou/gopsutil/cpu"
)

// MiscReporter implements the ReporterFunc interface
type MiscReporter struct {
	startTime time.Time
}

// NewMiscReporter initializes a MiscReporter object
func NewMiscReporter() MiscReporter {
	return MiscReporter{startTime: time.Now().UTC()}
}

// Report implements the Reporter interface it is called by the metrics server
func (m MiscReporter) Report(jWriter *flatjson.Writer) error {
	memValues, err := memvalues.GetMemValues()
	if err != nil {
		return errors.Wrap(err, "memvalues.GetMemValues()")
	}

	var cpuPercent []float64
	cpuPercent, err = cpu.Percent(time.Second, false)
	if err != nil {
		return errors.Wrap(err, "cpu.Percent failed")
	}

	for _, x := range []struct {
		key   string
		value interface{}
	}{
		{"system/start_time", duration2ms(time.Duration(m.startTime.UnixNano()))},
		{"system/cpu.pct", cpuPercent[0]},
		{"system/cpu_cores", runtime.NumCPU()},
		{"os", runtime.GOOS},
		{"os_arch", runtime.GOARCH},
		{"system/memory/available", memValues.SystemMemoryAvailable},
		{"system/memory/used", memValues.SystemMemoryUsed},
		{"system/memory/used_percent", memValues.SystemMemoryUsedPercent},
		{"process/memory/used", memValues.ProcessMemoryUsed},
	} {
		if err = jWriter.Write(x.key, x.value); err != nil {
			return errors.Wrap(err, "jWriter.Write")
		}
	}

	return nil
}

func duration2ms(d time.Duration) int64 {
	const nsPerMs = 1000000

	return d.Nanoseconds() / nsPerMs
}
