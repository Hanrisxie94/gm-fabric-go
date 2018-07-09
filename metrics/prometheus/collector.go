package prometheus

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	prom "github.com/prometheus/client_golang/prometheus"
)

// AllMetricsKey is a metrics key for the total of alll observations
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

// Collector collects Prometheus stats
type Collector interface {
	Collect(CollectorEntry) error
}

// CollectorType implements the Collector interface
type CollectorType struct {
	requestDurationVec *prom.SummaryVec
	requestSizeVec     *prom.CounterVec
	responseSizeVec    *prom.CounterVec
	tlsCount           prom.Counter
	nonTLSCount        prom.Counter
}

// NewCollector returrns an object that implements the Collector interface
func NewCollector() (*CollectorType, error) {
	var collector CollectorType

	// Objectives defines the quantile rank estimates with their respective
	// absolute error. If Objectives[q] = e, then the value reported for q
	// will be the φ-quantile value for some φ between q-e and q+e.
	//
	// This map of objectives is chosen to duplicate the dashboard metrics
	objectives := setObjectives()

	collector.requestDurationVec = prom.NewSummaryVec(
		prom.SummaryOpts{
			Name:       "http_request_duration_seconds",
			Help:       "duration of a single http request",
			Objectives: objectives,
		},
		LabelNames,
	)

	collector.requestSizeVec = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "http_request_size_bytes",
			Help: "number of bytes read from the request",
		},
		LabelNames,
	)

	collector.responseSizeVec = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "http_response_size_bytes",
			Help: "number of bytes written to the response",
		},
		LabelNames,
	)

	collector.tlsCount = prom.NewCounter(prom.CounterOpts{
		Name: "tls_requests",
		Help: "Number of requests using TLS.",
	})

	collector.nonTLSCount = prom.NewCounter(prom.CounterOpts{
		Name: "non_tls_requests",
		Help: "Number of requests not using TLS.",
	})

	for i, collector := range []prom.Collector{
		collector.requestDurationVec,
		collector.requestSizeVec,
		collector.responseSizeVec,
		collector.tlsCount,
		collector.nonTLSCount,
	} {
		if err := prom.Register(collector); err != nil {
			return nil, errors.Wrapf(err, "#%d:prometheus.Register", i)
		}
	}

	return &collector, nil
}

// Collect statistics by sending them to Prometheus
func (c *CollectorType) Collect(entry CollectorEntry) error {
	elapsed := computeElapsed(entry.StartTime, entry.EndTime)
	if elapsed == 0 {
		return nil
	}
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

	return nil
}

func setObjectives() map[float64]float64 {
	return map[float64]float64{
		0.5:    0.05,
		0.9:    0.01,
		0.95:   0.001,
		0.99:   0.001,
		0.999:  0.0001,
		0.9999: 0.00001,
	}
}

func computeElapsed(startTime, endTime time.Time) time.Duration {
	elapsed := endTime.Sub(startTime)
	if elapsed < 0 {
		elapsed = 0
	}

	return elapsed
}
