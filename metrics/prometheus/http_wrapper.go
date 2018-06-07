package prometheus

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	client "github.com/prometheus/client_golang/prometheus"

	"github.com/deciphernow/gm-fabric-go/metrics/httpmetrics"
)

var (
	// LabelNames is the list of valid lable names for this metric
	LabelNames = []string{"key", "method", "status"}

	// DefaultKeyFunc returns the URI as the metrics key
	DefaultKeyFunc KeyFunc = func(req *http.Request) string {
		return req.URL.EscapedPath()
	}
)

// KeyFunc geneates a histogram key from information in the http request
type KeyFunc func(*http.Request) string

type metricsState struct {
	hvec *client.HistogramVec
	rvec *client.CounterVec
}

// HandlerFactory wraps an http.Handler (inner) and captures metrics
type HandlerFactory interface {
	NewHandler(inner http.Handler) (http.Handler, error)
}

// NewHandlerFactory returns an abject that implements the HandlerFactory interface
// it is for use in creating
func NewHandlerFactory(
	firstBucket float64,
	bucketWidth float64,
	bucketCount int,
) (HandlerFactory, error) {
	var state metricsState

	state.hvec = client.NewHistogramVec(
		client.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "duration of a single http request",
			Buckets: client.LinearBuckets(firstBucket, bucketWidth, bucketCount),
		},
		LabelNames,
	)
	if err := client.Register(state.hvec); err != nil {
		return nil, errors.Wrap(err, "prometheus.Register hvec")
	}

	state.rvec = client.NewCounterVec(
		client.CounterOpts{
			Name: "http_response_size_bytes",
			Help: "number of bytes written to the response",
		},
		LabelNames,
	)
	if err := client.Register(state.rvec); err != nil {
		return nil, errors.Wrap(err, "prometheus.Register rvec")
	}

	return &state, nil
}

type handlerState struct {
	*metricsState
	keyFunc KeyFunc
	inner   http.Handler
}

func (mState *metricsState) NewHandlerWithKeyFunc(
	keyFunc KeyFunc,
	inner http.Handler,
) (http.Handler, error) {
	// TODO: I think we could curry the HistogramVec here
	var hState handlerState
	hState.metricsState = mState
	hState.keyFunc = keyFunc
	hState.inner = inner

	return &hState, nil
}

func (mState *metricsState) NewHandler(
	inner http.Handler,
) (http.Handler, error) {
	return mState.NewHandlerWithKeyFunc(DefaultKeyFunc, inner)
}

// ServeHTTP implements the http.Handler interface
func (hState *handlerState) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := httpmetrics.CountWriter{Next: w}
	startTime := time.Now()
	hState.inner.ServeHTTP(&c, req)
	endTime := time.Now()

	elapsed := endTime.Sub(startTime)
	if elapsed < 0 {
		elapsed = 0
	}

	method := strings.ToUpper(req.Method)
	if method == "" {
		method = "GET"
	}

	labels := client.Labels{
		"key":    hState.keyFunc(req),
		"method": method,
		"status": fmt.Sprintf("%d", c.Status),
	}

	hv, err := hState.hvec.GetMetricWith(labels)
	if err != nil {
		log.Printf("hState.hvec.GetMetricWith(%s) failed: %s", labels, err)
		return
	}
	hv.Observe(elapsed.Seconds())

	rv, err := hState.rvec.GetMetricWith(labels)
	if err != nil {
		log.Printf("hState.rvec.GetMetricWith(%s) failed: %s", labels, err)
		return
	}
	rv.Add(float64(c.BytesWritten))

}
