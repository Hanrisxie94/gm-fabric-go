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
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	prom "github.com/prometheus/client_golang/prometheus"

	"github.com/deciphernow/gm-fabric-go/metrics/httpmetrics"
)

// AllMetricsKey is a metrics key for the total of alll observations
const AllMetricsKey = "all"

type summaryMetricsState struct {
	requestDurationVec *prom.SummaryVec
	requestSizeVec     *prom.CounterVec
	responseSizeVec    *prom.CounterVec
}

// SummaryHandlerFactory wraps an http.Handler (inner) and captures metrics
type SummaryHandlerFactory interface {
	NewHandler(inner http.Handler) (http.Handler, error)
}

// NewSummaryHandlerFactory returns an object that implements the
// SummaryHandlerFactory interface
// it is for use in creating individual http.Handlers that are instrumented
// to collect our metrics.
func NewSummaryHandlerFactory() (SummaryHandlerFactory, error) {
	var state summaryMetricsState

	// Objectives defines the quantile rank estimates with their respective
	// absolute error. If Objectives[q] = e, then the value reported for q
	// will be the φ-quantile value for some φ between q-e and q+e.
	//
	// This map of objectives is chosen to duplicate the dashboard metrics
	objectives := map[float64]float64{
		0.5:    0.05,
		0.9:    0.01,
		0.95:   0.001,
		0.99:   0.001,
		0.999:  0.0001,
		0.9999: 0.00001,
	}

	state.requestDurationVec = prom.NewSummaryVec(
		prom.SummaryOpts{
			Name:       "http_request_duration_seconds",
			Help:       "duration of a single http request",
			Objectives: objectives,
		},
		LabelNames,
	)
	if err := prom.Register(state.requestDurationVec); err != nil {
		return nil, errors.Wrap(err, "prometheus.Register requestDurationVec")
	}

	state.requestSizeVec = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "http_request_size_bytes",
			Help: "number of bytes read from the request",
		},
		LabelNames,
	)
	if err := prom.Register(state.requestSizeVec); err != nil {
		return nil, errors.Wrap(err, "prometheus.Register requestSizeVec")
	}

	state.responseSizeVec = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "http_response_size_bytes",
			Help: "number of bytes written to the response",
		},
		LabelNames,
	)
	if err := prom.Register(state.responseSizeVec); err != nil {
		return nil, errors.Wrap(err, "prometheus.Register responseSizeVec")
	}

	return &state, nil
}

type summaryHandlerState struct {
	*summaryMetricsState
	keyFunc HTTPKeyFunc
	inner   http.Handler
}

// NewHandlerWithKeyFunc creates a new http.Handler instrumented to collect
// our metrics.
// With a specilaized key function.
func (mState *summaryMetricsState) NewHandlerWithKeyFunc(
	keyFunc HTTPKeyFunc,
	inner http.Handler,
) (http.Handler, error) {
	var hState summaryHandlerState
	hState.summaryMetricsState = mState
	hState.keyFunc = keyFunc
	hState.inner = inner

	return &hState, nil
}

// NewHandler creates a new http.Handler instrumented to collect our metrics
// NewHandler uses DefaultKeyFunc
func (mState *summaryMetricsState) NewHandler(
	inner http.Handler,
) (http.Handler, error) {
	return mState.NewHandlerWithKeyFunc(DefaultHTTPKeyFunc, inner)
}

// ServeHTTP implements the http.Handler interface
// It collects:
//      http_request_duration_seconds
//      http_request_size_bytes
//      http_response_size_bytes
func (hState *summaryHandlerState) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	responseWriter := httpmetrics.CountWriter{Next: w}

	requestReader := httpmetrics.CountReader{Next: req.Body}
	req.Body = &requestReader

	startTime := time.Now()
	hState.inner.ServeHTTP(&responseWriter, req)
	endTime := time.Now()

	elapsed := endTime.Sub(startTime)
	if elapsed < 0 {
		elapsed = 0
	}

	method := strings.ToUpper(req.Method)
	if method == "" {
		method = "GET"
	}

	status := responseWriter.Status
	if status == 0 {
		status = 200
	}

	for _, labels := range []prom.Labels{
		prom.Labels{
			"key":    hState.keyFunc(req),
			"method": method,
			"status": fmt.Sprintf("%d", status),
		},
		prom.Labels{
			"key":    AllMetricsKey,
			"method": "",
			"status": fmt.Sprintf("%d", status),
		},
	} {
		requestDuration, err := hState.requestDurationVec.GetMetricWith(labels)
		if err != nil {
			log.Printf("hState.requestDurationVec.GetMetricWith(%s) failed: %s", labels, err)
			return
		}
		requestDuration.Observe(elapsed.Seconds())
		requestSize, err := hState.requestSizeVec.GetMetricWith(labels)
		if err != nil {
			log.Printf("hState.requestSizeVec.GetMetricWith(%s) failed: %s", labels, err)
			return
		}
		requestSize.Add(float64(requestReader.BytesRead))
		responseSize, err := hState.responseSizeVec.GetMetricWith(labels)
		if err != nil {
			log.Printf("hState.responseSizeVec.GetMetricWith(%s) failed: %s", labels, err)
			return
		}
		responseSize.Add(float64(responseWriter.BytesWritten))
	}

}
