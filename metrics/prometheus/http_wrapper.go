package prometheus

import (
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"

	client "github.com/prometheus/client_golang/prometheus"
)

var (
	// LabelNames is the list of valid lable names for this metric
	LabelNames = []string{"key", "status"}
)

type metricsState struct {
	hvec *client.HistogramVec
}

// HandlerFactory wraps an http.Handler (inner) and captures metrics
type HandlerFactory interface {
	NewHandler(key string, inner http.Handler) (http.Handler, error)
}

// NewHandlerFactory returns an abject that implements the HandlerFactory interface
// it is for use in creating
func NewHandlerFactory(
	firstBucket float64,
	bucketWidth float64,
	bucketCount int,
) (HandlerFactory, error) {
	state := metricsState{
		hvec: client.NewHistogramVec(
			client.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "duration of a single http request",
				Buckets: client.LinearBuckets(firstBucket, bucketWidth, bucketCount),
			},
			LabelNames,
		),
	}

	if err := client.Register(state.hvec); err != nil {
		return nil, errors.Wrap(err, "prometheus.Register")
	}

	return &state, nil
}

type handlerState struct {
	*metricsState
	inner http.Handler
}

func (mState *metricsState) NewHandler(
	key string,
	inner http.Handler,
) (http.Handler, error) {
	// TODO: I think we could curry the HistogramVec here
	var hState handlerState
	hState.metricsState = mState
	hState.inner = inner

	return &hState, nil
}

// ServeHTTP implements the http.Handler interface
func (hState *handlerState) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now()
	hState.inner.ServeHTTP(w, req)
	endTime := time.Now()

	elapsed := endTime.Sub(startTime)
	if elapsed < 0 {
		elapsed = 0
	}

	labels := client.Labels{
		"key":    req.URL.EscapedPath(),
		"status": "200",
	}

	observer, err := hState.hvec.GetMetricWith(labels)
	if err != nil {
		log.Printf("hState.hvec.GetMetricWith(%s) failed: %s", labels, err)
		return
	}

	observer.Observe(elapsed.Seconds())
}
