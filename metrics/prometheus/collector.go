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
	"fmt"
	"time"

	"github.com/pkg/errors"
	prom "github.com/prometheus/client_golang/prometheus"
)

// AllMetricsKey is a metrics key for the total of  observations
const AllMetricsKey = "all"

// CollectorEntry contains the stats data for one transaction
type CollectorEntry struct {
	StartTime    time.Time
	EndTime      time.Time
	Method       string
	Status       int
	Key          string
	BytesRead    uint64
	BytesWritten uint64
	TLS          bool
	Err          error
}

// SystemMetricsEntry contains system metrics, reported periodically
type SystemMetricsEntry struct {
	SystemCPUPercent        float64 // "system_cpu_pct"
	SystemCPUCores          float64 // "system_cpu_cores"
	SystemMemoryAvailable   float64 // "system_memory_available"
	SystemMemoryUsed        float64 // "system_memory_used"
	SystemMemoryUsedPercent float64 // "system_memory_used_percent"
	ProcessMemoryUsed       float64 // "process_memory_used"
}

// Collector collects Prometheus stats
type Collector interface {
	Collect(CollectorEntry) error
	CollectSystemMetrics(SystemMetricsEntry)
}

// CollectorType implements the Collector interface
type CollectorType struct {
	requestDurationVec           *prom.HistogramVec
	requestSizeVec               *prom.CounterVec
	responseSizeVec              *prom.CounterVec
	tlsCount                     prom.Counter
	nonTLSCount                  prom.Counter
	systemStartTimeGauge         prom.Gauge
	systemCPUPercentGauge        prom.Gauge
	systemCPUCoresGauge          prom.Gauge
	systemMemoryAvailableGauge   prom.Gauge
	systemMemoryUsedGauge        prom.Gauge
	systemMemoryUsedPercentGauge prom.Gauge
	processMemoryUsedGauge       prom.Gauge
}

// NewCollector returns an object that implements the Collector interface
func NewCollector() (*CollectorType, error) {
	collector := CollectorType{
		requestDurationVec:           createRequestDurationHistogram(),
		requestSizeVec:               createRequestSizeVector(),
		responseSizeVec:              createResponseSizeVector(),
		tlsCount:                     createTLSCounter(),
		nonTLSCount:                  createNonTLSCounter(),
		systemStartTimeGauge:         createSystemStartTimeGauge(),
		systemCPUPercentGauge:        createSystemCPUPercentGauge(),
		systemCPUCoresGauge:          createSystemCPUCoresGauge(),
		systemMemoryAvailableGauge:   createSystemMemoryAvailableGauge(),
		systemMemoryUsedGauge:        createSystemMemoryUsedGauge(),
		systemMemoryUsedPercentGauge: createSystemMemoryUsedPercentGauge(),
		processMemoryUsedGauge:       createProcessMemoryUsedGauge(),
	}

	for i, c := range []prom.Collector{
		collector.requestDurationVec,
		collector.requestSizeVec,
		collector.responseSizeVec,
		collector.tlsCount,
		collector.nonTLSCount,
		collector.systemStartTimeGauge,
		collector.systemCPUPercentGauge,
		collector.systemCPUCoresGauge,
		collector.systemMemoryAvailableGauge,
		collector.systemMemoryUsedGauge,
		collector.systemMemoryUsedPercentGauge,
		collector.processMemoryUsedGauge,
	} {
		if err := prom.Register(c); err != nil {
			return nil, errors.Wrapf(err, "#%d:prometheus.Register", i)
		}
	}

	collector.systemStartTimeGauge.SetToCurrentTime()

	return &collector, nil
}

func createRequestDurationHistogram() *prom.HistogramVec {
	// from github.com/prometheus/client_golang/prometheus/histogram.go
	// see also LinearBuckets and ExponentialBuckets in the same file
	//
	// DefBuckets are the default Histogram buckets. The default buckets are
	// tailored to broadly measure the response time (in seconds) of a network
	// service. Most likely, however, you will be required to define buckets
	// customized to your use case.
	//
	//	DefBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
	//

	return prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "duration of a single http request",
			Buckets: prom.DefBuckets,
		},
		LabelNames,
	)
}

func createRequestSizeVector() *prom.CounterVec {
	return prom.NewCounterVec(
		prom.CounterOpts{
			Name: "http_request_size_bytes",
			Help: "number of bytes read from the request",
		},
		LabelNames,
	)
}

func createResponseSizeVector() *prom.CounterVec {
	return prom.NewCounterVec(
		prom.CounterOpts{
			Name: "http_response_size_bytes",
			Help: "number of bytes written to the response",
		},
		LabelNames,
	)
}

func createTLSCounter() prom.Counter {
	return prom.NewCounter(prom.CounterOpts{
		Name: "tls_requests",
		Help: "Number of requests using TLS.",
	})
}

func createNonTLSCounter() prom.Counter {
	return prom.NewCounter(prom.CounterOpts{
		Name: "non_tls_requests",
		Help: "Number of requests not using TLS.",
	})
}

func createSystemStartTimeGauge() prom.Gauge {
	return prom.NewGauge(prom.GaugeOpts{
		Name: "system_start_time_seconds",
		Help: "The time the system started running.",
	})
}

func createSystemCPUPercentGauge() prom.Gauge {
	return prom.NewGauge(prom.GaugeOpts{
		Name: "system_cpu_pct",
		Help: "Percent of CPU time in use by the system.",
	})
}

func createSystemCPUCoresGauge() prom.Gauge {
	return prom.NewGauge(prom.GaugeOpts{
		Name: "system_cpu_cores",
		Help: "The number of CPU cores avaialble.",
	})
}

func createSystemMemoryAvailableGauge() prom.Gauge {
	return prom.NewGauge(prom.GaugeOpts{
		Name: "system_memory_available",
		Help: "The amount of memory available on the system.",
	})
}

func createSystemMemoryUsedGauge() prom.Gauge {
	return prom.NewGauge(prom.GaugeOpts{
		Name: "system_memory_used",
		Help: "The amount of memory currently used by the system.",
	})
}

func createSystemMemoryUsedPercentGauge() prom.Gauge {
	return prom.NewGauge(prom.GaugeOpts{
		Name: "system_memory_used_percent",
		Help: "The of percentage of available memory currently used by the system.",
	})
}

func createProcessMemoryUsedGauge() prom.Gauge {
	return prom.NewGauge(prom.GaugeOpts{
		Name: "process_memory_used",
		Help: "The amount of memory currently used by this process.",
	})
}

// Collect statistics by sending them to Prometheus
func (c *CollectorType) Collect(entry CollectorEntry) error {
	elapsed := computeElapsed(entry.StartTime, entry.EndTime)
	if elapsed > 0 {
		for _, labels := range []prom.Labels{
			prom.Labels{
				"key":    entry.Key,
				"method": entry.Method,
				"status": fmt.Sprintf("%d", entry.Status),
			},
			prom.Labels{
				"key":    AllMetricsKey,
				"method": "",
				"status": fmt.Sprintf("%d", entry.Status),
			},
		} {
			requestDuration, err := c.requestDurationVec.GetMetricWith(labels)
			if err != nil {
				return errors.Wrapf(err, "requestDurationVec.GetMetricWith(%s)", labels)
			}
			requestDuration.Observe(elapsed.Seconds())
			requestSize, err := c.requestSizeVec.GetMetricWith(labels)
			if err != nil {
				return errors.Wrapf(err, "requestSizeVec.GetMetricWith(%s)", labels)
			}
			requestSize.Add(float64(entry.BytesRead))
			responseSize, err := c.responseSizeVec.GetMetricWith(labels)
			if err != nil {
				return errors.Wrapf(err, "responseSizeVec.GetMetricWith(%s)", labels)
			}
			responseSize.Add(float64(entry.BytesWritten))
		}

		if entry.TLS {
			c.tlsCount.Inc()
		} else {
			c.nonTLSCount.Inc()
		}
	}

	return nil
}

// CollectSystemMetrics sends system metrics to Prometheus
func (c *CollectorType) CollectSystemMetrics(entry SystemMetricsEntry) {
	for _, d := range []struct {
		gauge prom.Gauge
		value float64
	}{
		{c.systemCPUPercentGauge, entry.SystemCPUPercent},
		{c.systemCPUCoresGauge, entry.SystemCPUCores},
		{c.systemMemoryAvailableGauge, entry.SystemMemoryAvailable},
		{c.systemMemoryUsedGauge, entry.SystemMemoryUsed},
		{c.systemMemoryUsedPercentGauge, entry.SystemMemoryUsedPercent},
		{c.processMemoryUsedGauge, entry.ProcessMemoryUsed},
	} {
		d.gauge.Set(d.value)
	}
}

func computeElapsed(startTime, endTime time.Time) time.Duration {
	elapsed := endTime.Sub(startTime)
	if elapsed < 0 {
		elapsed = 0
	}

	return elapsed
}
